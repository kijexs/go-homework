package usecase

import (
	"context"
	"go-homework/internal/domain"
)

type MovieFetcher func(ctx context.Context, id int) (domain.Movie, error)

func worker(ctx context.Context, fetch MovieFetcher, jobs <-chan int, results chan<- domain.Result) {
	for movieId := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			movie, err := fetch(ctx, movieId)

			result := domain.Result{
				Id:   movieId,
				Data: movie,
				Err:  err,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
