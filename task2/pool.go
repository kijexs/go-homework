package main

import (
	"context"
	"net/http"
	"sync"
	"fmt"
)

func run(ctx context.Context, cfg Config) {
	var httpClient = &http.Client{Timeout: cfg.Timeout}

	var wg sync.WaitGroup

	jobs := make(chan int)
	results := make(chan Result)

	go func() {
		defer close(jobs)
		for i := cfg.From; i <= cfg.To; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}

		}
	}()

	for i := 0; i < cfg.Workers; i++ {
		wg.Go(func() {
			worker(ctx, httpClient, cfg.Timeout, jobs, results)
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		fmt.Println(res)
	}
}
