package aquifer

import (
	"time"
)

type Conn interface {
	Close() error
}

type connWrapper[C Conn] struct {
	conn       C
	createdAt  time.Time
	lastUsedAt time.Time
	useCount   int64
}

func newConnWrapper[C Conn](conn C) *connWrapper[C] {
	return &connWrapper[C]{
		conn:       conn,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
	}
}

func (c *connWrapper[C]) isExpired(idleTimeout time.Duration) bool {
	return time.Since(c.lastUsedAt) > idleTimeout
}

func (c *connWrapper[C]) markUsed() {
	c.lastUsedAt = time.Now()
	c.useCount++
}
