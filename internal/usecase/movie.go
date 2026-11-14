package usecase

import (
	"context"
	"fmt"

	"go-homework/internal/domain"
)

type MovieRepository interface {
	GetByID(ctx context.Context, id int) (domain.Movie, error)
	Save(ctx context.Context, id int, movie domain.Movie) error
	GetLast(ctx context.Context) (domain.Result, error)
	GetNotExistingIDs(ctx context.Context, maxId int) ([]int, error)
}

type MovieService struct {
	repository  MovieRepository
	workerCount int
}

func NewMovieService(repository MovieRepository, workerCount int) *MovieService {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &MovieService{
		repository:  repository,
		workerCount: workerCount,
	}
}

func (s *MovieService) LoadMovies(ctx context.Context, req domain.MovieRequest) ([]domain.Result, error) {
	if req.From <= 0 || req.To <= 0 || req.From > req.To {
		return nil, fmt.Errorf("Load movies: %w", domain.ErrInvalidArgument)
	}

	ids := make([]int, 0, req.To-req.From)
	for id := req.From; id <= req.To; id++ {
		ids = append(ids, id)
	}
	return loadMoviesParallel(ctx, ids, s.workerCount, s.repository.GetByID)
}

func (s *MovieService) GetLastMovie(ctx context.Context) (domain.Result, error) {
	return s.repository.GetLast(ctx)
}
