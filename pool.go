package aquifer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Pool[C Conn] struct {
	cfg     *config
	dial    func(ctx context.Context) (C, error)
	idle    chan *connWrapper[C]
	open    atomic.Int64
	inUse   atomic.Int64
	waiters atomic.Int64
	mu      sync.Mutex
	closed  bool
	done    chan struct{}
}

type Stats struct {
	Open    int64
	Idle    int64
	InUse   int64
	Waiters int64
}

func (p *Pool[C]) dialOne(ctx context.Context) (*connWrapper[C], error) {
	var zero *connWrapper[C]
	ctx, cancel := context.WithTimeout(ctx, p.cfg.dialTimeout)
	defer cancel()

	conn, err := p.dial(ctx)
	if err != nil {
		return zero, &DialError{Err: err}
	}
	p.open.Add(1)

	return newConnWrapper(conn), nil
}

func New[C Conn](dial func(ctx context.Context) (C, error), opts ...Option) (*Pool[C], error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	p := &Pool[C]{
		cfg:  cfg,
		dial: dial,
		idle: make(chan *connWrapper[C], cfg.maxConns),
		done: make(chan struct{}),
	}

	for i := 0; i < cfg.minConns; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.dialTimeout)
		conn, err := p.dialOne(ctx)
		cancel()
		if err != nil {
			for len(p.idle) > 0 {
				c := <-p.idle
				_ = c.conn.Close()
			}
			return nil, fmt.Errorf("aquifer: pre-warm failed: %w", err)
		}
		p.idle <- conn
	}

	if cfg.idleTimeout > 0 {
		go p.reaper()
	}

	return p, nil
}

func (p *Pool[C]) Acquire(ctx context.Context) (C, error) {
	var zero C

	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return zero, ErrClosed
	}

	select {
	case w := <-p.idle:
		p.inUse.Add(1)
		w.markUsed()
		return w.conn, nil
	default:
	}

	if p.open.Load() < int64(p.cfg.maxConns) {
		c, err := p.dialOne(ctx)
		if err != nil {
			return zero, fmt.Errorf("aquifer: dial: %w", err)
		}
		p.inUse.Add(1)
		return c.conn, nil
	}

	p.waiters.Add(1)
	defer p.waiters.Add(-1)

	select {
	case w := <-p.idle:
		p.inUse.Add(1)
		w.markUsed()
		return w.conn, nil
	case <-ctx.Done():
		return zero, ErrExhausted
	case <-p.done:
		return zero, ErrClosed
	}
}

func (p *Pool[C]) Release(conn C) {
	p.inUse.Add(-1)
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		_ = conn.Close()
		p.open.Add(-1)
		return
	}

	w := newConnWrapper(conn)

	select {
	case p.idle <- w:
	default:
		_ = conn.Close()
		p.open.Add(-1)
	}
}

func (p *Pool[C]) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}

	p.closed = true
	close(p.done)
	p.mu.Unlock()

	for len(p.idle) > 0 {
		w := <-p.idle
		_ = w.conn.Close()
		p.open.Add(-1)
	}
}

func (p *Pool[C]) Stats() Stats {
	return Stats{
		Idle:    int64(len(p.idle)),
		Open:    p.open.Load(),
		InUse:   p.inUse.Load(),
		Waiters: p.waiters.Load(),
	}
}
