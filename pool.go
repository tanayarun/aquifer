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
	idle    chan C
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

func (p *Pool[C]) dialOne(ctx context.Context) (C, error) {
	var zero C
	ctx, cancel := context.WithTimeout(ctx, p.cfg.dialTimeout)
	defer cancel()

	conn, err := p.dial(ctx)
	if err != nil {
		return zero, &DialError{Err: err}
	}
	p.open.Add(1)

	return conn, nil
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
		idle: make(chan C, cfg.maxConns),
		done: make(chan struct{}),
	}

	for i := 0; i < cfg.minConns; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.dialTimeout)
		conn, err := p.dialOne(ctx)
		cancel()
		if err != nil {
			for len(p.idle) > 0 {
				c := <-p.idle
				c.Close()
			}
			return nil, fmt.Errorf("aquifer: pre-warm failed: %w", err)
		}
		p.idle <- conn
	}

	return p, nil
}
