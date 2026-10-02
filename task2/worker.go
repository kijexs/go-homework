package main

import (
	"context"
	"net/http"
	"time"
)

func worker(ctx context.Context, client *http.Client, timeout time.Duration, jobs <-chan int, results chan<- Result) {
	for movieId := range jobs {
		select {
		case <-ctx.Done():
			return
		case results <- fetchMovie(ctx, client, timeout, movieId):

		}
	}
}
