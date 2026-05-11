package aquifer

import (
	"context"
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
