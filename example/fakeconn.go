package main

import (
	"fmt"
	"sync"
)

type fakeConn struct {
	id     int
	closed bool
	mu     sync.Mutex
}

func (f *fakeConn) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return fmt.Errorf("fakeconn %d: already closed", f.id)
	}
	f.closed = true
	fmt.Printf("conn %d closed\n", f.id)

	return nil
}
