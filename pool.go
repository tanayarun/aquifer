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
	Inuse   int64
	Waiters int64
}
