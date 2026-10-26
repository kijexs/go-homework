package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"strconv"
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

type Stringer interface {
    String() string
}

func (res Result) String() string {
	switch {
	case res.Err == nil:
		return fmt.Sprintf(
			"id: %d | year: %d | title: %s | director: %s\n",
		 	res.Id, res.Data.Year, res.Data.Title, res.Data.Director,
		)
		
	case errors.Is(res.Err, context.Canceled):
		return fmt.Sprintf("id: %d | canceled: %v\n", res.Id, res.Err)

	case errors.Is(res.Err, context.DeadlineExceeded):
		return fmt.Sprintf("id: %d | deadline exceeded: %v\n", res.Id, res.Err)

	default:
		return fmt.Sprintf("id: %d | error: %v\n", res.Id, res.Err)
	}
}

func fetchMovie(ctx context.Context, client *http.Client, timeout time.Duration, movieId int) Result {
	result := Result{Id: movieId}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var req, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://homeworksite.site/"+ strconv.Itoa(movieId) +"/info.0.json",
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
