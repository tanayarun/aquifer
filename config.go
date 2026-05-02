package aquifer

import (
	"fmt"
	"time"
)

type config struct {
	minConns    int
	maxConns    int
	idleTimeout time.Duration
	dialTimeout time.Duration
}

type Option func(*config)

func defaultConfig() *config {
	return &config{
		minConns:    1,
		maxConns:    5,
		idleTimeout: 5 * time.Minute,
		dialTimeout: 10 * time.Second,
	}
}

func WithMinConns(n int) Option {
	return func(c *config) {
		c.minConns = n
	}
}

func WithMaxConns(n int) Option {
	return func(c *config) {
		c.maxConns = n
	}
}

func WithIdleTimeout(d time.Duration) Option {
	return func(c *config) {
		c.idleTimeout = d
	}
}

func WithDialTimeout(d time.Duration) Option {
	return func(c *config) {
		c.dialTimeout = d
	}
}

func (c *config) validate() error {
	if c.maxConns < 1 {
		return fmt.Errorf("aquifer: maxConns must be >= 1")
	}
	if c.minConns > c.maxConns {
		return fmt.Errorf("aquifer: minConns (%d) cannot exceed maxConns (%d)", c.minConns, c.maxConns)
	}
	return nil
}
