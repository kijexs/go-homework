package usecase

import (
	"context"
	"go-homework/internal/domain"
)

func worker(ctx context.Context, service *MovieService, jobs <-chan int, results chan<- domain.Result) {
	for movieId := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			movie, err := service.provider.GetMovie(ctx, movieId)

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
