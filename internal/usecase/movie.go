package usecase

import (
	"context"
	"fmt"

	"go-homework/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieService struct {
	provider    MovieProvider
	workerCount int
}

func NewMovieService(provider MovieProvider, workerCount int) *MovieService {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &MovieService{
		provider:    provider,
		workerCount: workerCount,
	}
}

func (s *MovieService) LoadMovies(ctx context.Context, req domain.MovieRequest) ([]domain.Result, error) {
	if req.From <= 0 || req.To <= 0 || req.From > req.To {
		return nil, fmt.Errorf("Load movies: %w", domain.ErrInvalidArgument)
	}

	return s.loadMoviesParallel(ctx, req)
}

func (s *MovieService) GetLastMovie(ctx context.Context) (domain.Result, error) {
	low, high, err := s.findUpperBound(ctx)
	if err != nil {
		return domain.Result{}, err
	}

	return s.binarySearchLastMovie(ctx, low, high)
}
