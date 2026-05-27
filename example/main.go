package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tanayarun/aquifer"
)

func main() {
	var counter int

	dial := func(ctx context.Context) (*fakeConn, error) {
		counter++
		fmt.Printf("dialling conn %d\n", counter)
		return &fakeConn{id: counter}, nil
	}

	pool, err := aquifer.New(dial,
		aquifer.WithMinConns(2),
		aquifer.WithMaxConns(5),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Printf("stats: %+v\n", pool.Stats())

	var wg sync.WaitGroup

	for range 3 {
		wg.Go(func() {
			conn, err := pool.Acquire(context.Background())
			if err != nil {
				fmt.Printf("acquire error: %v\n", err)
				return
			}
			defer pool.Release(conn)

			fmt.Printf("acquired conn %d — stats: %+v\n", conn.id, pool.Stats())
			time.Sleep(100 * time.Millisecond)
		})
	}

	wg.Wait()
	fmt.Printf("final stats: %+v\n", pool.Stats())
}
