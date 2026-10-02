package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Movie struct {
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type Result struct {
	Id   int
	Data Movie
	Err  error
}

func fetchMovie(ctx context.Context, client *http.Client, timeout time.Duration, movieId int) Result {
	result := Result{Id: movieId}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	id := fmt.Sprintf("%d", movieId)
	var req, err = http.NewRequestWithContext(
		ctx,
		"GET",
		"https://homeworksite.site/"+id+"/info.0.json",
		nil,
	)
	if err != nil {
		result.Err = fmt.Errorf("error forming a new request: %w", err)
		return result
	}
	res, err := client.Do(req)
	if err != nil {
		result.Err = fmt.Errorf("error executing request: %w", err)
		return result
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		result.Err = fmt.Errorf("unexpected status: got %v", res.StatusCode)
		return result
	}

	var movie Movie
	err = json.NewDecoder(res.Body).Decode(&movie)
	if err != nil {
		result.Err = fmt.Errorf("decode response: %w", err)
		return result
	}
	result.Data = movie

	return result
}

func printResult(res Result) {
	switch {
	case res.Err == nil:
		fmt.Printf("id: %d | year: %d | title: %s | director: %s\n", res.Id, res.Data.Year, res.Data.Title, res.Data.Director)
		return
	case errors.Is(res.Err, context.Canceled):
		fmt.Printf("id: %d | canceled: %v\n", res.Id, res.Err)
		return
	case errors.Is(res.Err, context.DeadlineExceeded):
		fmt.Printf("id: %d | deadline exceeded: %v\n", res.Id, res.Err)
		return
	default:
		fmt.Printf("id: %d | error: %v\n", res.Id, res.Err)
		return
	}
}
