package usecase

import (
	"context"
	"sync"

	"go-homework/internal/domain"
)

func (s *MovieService) loadMoviesParallel(ctx context.Context, req domain.MovieRequest) ([]domain.Result, error) {
	result := []domain.Result{}
	var wg sync.WaitGroup

	jobs := make(chan int)
	results := make(chan domain.Result)

	go func() {
		defer close(jobs)
		for i := req.From; i <= req.To; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}

		}
	}()

	for i := 0; i < s.workerCount; i++ {
		wg.Go(func() {
			worker(ctx, s, jobs, results)
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
