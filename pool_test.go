package aquifer_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tanayarun/aquifer"
)

type fakeConn struct{ closed bool }

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

func fakeDial() func(context.Context) (*fakeConn, error) {
	return func(_ context.Context) (*fakeConn, error) {
		return &fakeConn{}, nil
	}
}

func TestAcquireRelease(t *testing.T) {
	pool, err := aquifer.New(fakeDial(),
		aquifer.WithMinConns(1),
		aquifer.WithMaxConns(3),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pool.Close()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	if pool.Stats().InUse != 1 {
		t.Errorf("want InUse=1, got %d", pool.Stats().InUse)
	}

	pool.Release(conn)

	if pool.Stats().InUse != 0 {
		t.Errorf("want InUse=0, got %d", pool.Stats().InUse)
	}
}

func TestMaxConns(t *testing.T) {
	pool, err := aquifer.New(fakeDial(),
		aquifer.WithMinConns(1),
		aquifer.WithMaxConns(2),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pool.Close()

	conn1, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire conn1: %v", err)
	}

	conn2, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire conn2: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = pool.Acquire(ctx)
	if !errors.Is(err, aquifer.ErrExhausted) {
		t.Errorf("want ErrExhausted, got %v", err)
	}

	pool.Release(conn1)
	pool.Release(conn2)
}

func TestClosedPool(t *testing.T) {
	pool, err := aquifer.New(fakeDial(),
		aquifer.WithMinConns(1),
		aquifer.WithMaxConns(3),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	pool.Close()

	_, err = pool.Acquire(context.Background())
	if !errors.Is(err, aquifer.ErrClosed) {
		t.Errorf("want ErrClosed, got %v", err)
	}
}

func TestConcurrentAcquireRelease(t *testing.T) {
	pool, err := aquifer.New(fakeDial(),
		aquifer.WithMinConns(2),
		aquifer.WithMaxConns(5),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pool.Close()

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := pool.Acquire(context.Background())
			if err != nil {
				t.Errorf("Acquire: %v", err)
				return
			}
			time.Sleep(10 * time.Millisecond)
			pool.Release(conn)
		}()
	}

	wg.Wait()

	if pool.Stats().InUse != 0 {
		t.Errorf("want InUse=0 after all released, got %d", pool.Stats().InUse)
	}
}
