package aquifer_test

import (
	"context"
	"testing"
	"time"

	"github.com/tanayarun/aquifer"
)

func TestIsExpired(t *testing.T) {
	t.Log("isExpired tested indirectly via TestReaper")
}

func TestReaperEvicts(t *testing.T) {
	pool, err := aquifer.New(
		fakeDial(),
		aquifer.WithMinConns(1),
		aquifer.WithMaxConns(5),
		aquifer.WithIdleTimeout(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pool.Close()

	for range 3 {
		conn, _ := pool.Acquire(context.Background())
		pool.Release(conn)
	}

	before := pool.Stats().Open

	time.Sleep(200 * time.Millisecond)

	after := pool.Stats().Open

	if after >= before {
		t.Errorf("reaper should have evicted conns: before=%d after=%d", before, after)
	}
}
