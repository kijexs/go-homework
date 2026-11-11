package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go-homework/domain"
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
	result := []domain.Result{}
	var wg sync.WaitGroup

	jobs := make(chan int)
	results := make(chan domain.Result)

	go func() {
		defer close(jobs)
		for i := req.From; i <= req.To; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}

		}
	}()

	for i := 0; i < s.workerCount; i++ {
		wg.Go(func() {
			worker(ctx, s, jobs, results)
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		result = append(result, res)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MovieService) GetLastMovie(ctx context.Context) (domain.Result, error) {
	low := 1
	hight := 1

	for {
		_, err := s.provider.GetMovie(ctx, hight)

		if errors.Is(err, domain.ErrNotFound) {
			break
		}
		if err != nil {
			return domain.Result{}, err
		}
		low = hight
		hight *= 2
	}

	lastFilm := domain.Result{}

	for low <= hight {
		mid := low + ((hight - low) / 2)
		res, err := s.provider.GetMovie(ctx, mid)
		if err == nil {
			lastFilm = domain.Result{
				Id:   mid,
				Data: res,
				Err:  err,
			}
			low = mid + 1
		} else if errors.Is(err, domain.ErrNotFound) {
			hight = mid - 1
		} else {
			return domain.Result{}, err
		}
	}

	if lastFilm.Id == 0 {
		return domain.Result{}, fmt.Errorf("Get last movie: %w", domain.ErrNotFound)
	}

	return lastFilm, nil
}
