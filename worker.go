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
		default:
			result := fetchMovie(ctx, client, timeout, movieId)
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
