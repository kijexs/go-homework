package domain

import (
	"context"
)

type MovieRequest struct {
	To   int
	From int
}

type Movie struct {
	Title    string
	Year     int
	Director string
}

type Result struct {
	Id   int
	Data Movie
	Err  error
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, req MovieRequest) ([]Result, error)
	GetLastMovie(ctx context.Context) (Result, error)
}
