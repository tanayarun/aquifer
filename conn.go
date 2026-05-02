package aquifer

import (
	"time"
)

type Conn interface {
	Close() error
}

type connWrapper struct {
	conn       Conn
	createdAt  time.Time
	lastUsedAt time.Time
	useCount   int64
}

func newConnWrapper(conn Conn) *connWrapper {
	return &connWrapper{
		conn:       conn,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
	}
}

func (c *connWrapper) isExpired(idleTimeout time.Duration) bool {
	return time.Since(c.lastUsedAt) > idleTimeout
}

func (c *connWrapper) markUsed() {
	c.lastUsedAt = time.Now()
	c.useCount++
}
