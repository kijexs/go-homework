package movie

import (
	"context"
)

func worker(ctx context.Context, service *MovieService, jobs <-chan int, results chan<- Result) {
	for movieId := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			result := fetchMovie(ctx, *service, movieId)
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
