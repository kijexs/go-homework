package usecase

import (
	"context"
	"sync"

	"go-homework/internal/domain"
)

func loadMoviesParallel(ctx context.Context, ids []int, workerCount int, fetch MovieFetcher) ([]domain.Result, error) {
	result := []domain.Result{}
	var wg sync.WaitGroup

	jobs := make(chan int)
	results := make(chan domain.Result)

	go func() {
		defer close(jobs)
		for _, id := range ids {
			select {
			case <-ctx.Done():
				return
			case jobs <- id:
			}

		}
	}()

	for i := 0; i < workerCount; i++ {
		wg.Go(func() {
			worker(ctx, fetch, jobs, results)
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		result = append(result, res)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
