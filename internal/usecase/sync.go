package usecase

import (
	"context"
	"fmt"

	"go-homework/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieSyncService struct {
	repository  MovieRepository
	provider    MovieProvider
	workerCount int
}

func NewMovieSyncService(repository MovieRepository, provider MovieProvider, workerCount int) *MovieSyncService {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &MovieSyncService{
		repository:  repository,
		provider:    provider,
		workerCount: workerCount,
	}
}

func (s *MovieSyncService) Sync(ctx context.Context) error {
	low, high, err := s.findUpperBound(ctx)
	if err != nil {
		return fmt.Errorf("find upper bound: %w", err)
	}

	lastMovie, err := s.binarySearchLastMovie(ctx, low, high)
	if err != nil {
		return fmt.Errorf("find last movie: %w", err)
	}

	notExistingIDs, err := s.repository.GetNotExistingIDs(ctx, lastMovie.Id)
	if err != nil {
		return fmt.Errorf("get missing IDs: %w", err)
	}
	if len(notExistingIDs) == 0 {
		return nil
	}

	results, err := loadMoviesParallel(ctx, notExistingIDs, s.workerCount, s.provider.GetMovie)
	if err != nil {
		return fmt.Errorf("download movies: %w", err)
	}

	for _, result := range results {
		if result.Err != nil {
			return fmt.Errorf("download movie %d: %w", result.Id, result.Err)
		}
		if err := s.repository.Save(ctx, result.Id, result.Data); err != nil {
			return fmt.Errorf("save movie %d: %w", result.Id, err)
		}
	}

	return nil
}
