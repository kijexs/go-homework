package usecase

import (
	"context"
	"errors"
	"fmt"

	"go-homework/internal/domain"
)

func (s *MovieService) findUpperBound(ctx context.Context) (int, int, error) {
	low := 1
	high := 1

	for {
		_, err := s.provider.GetMovie(ctx, high)

		if errors.Is(err, domain.ErrNotFound) {
			break
		}
		if err != nil {
			return 0, 0, err
		}
		low = high
		high *= 2
	}
	return low, high, nil
}

func (s *MovieService) binarySearchLastMovie(ctx context.Context, low, high int) (domain.Result, error) {
	lastFilm := domain.Result{}

	for low <= high {
		mid := low + ((high - low) / 2)
		res, err := s.provider.GetMovie(ctx, mid)
		if err == nil {
			lastFilm = domain.Result{
				Id:   mid,
				Data: res,
			}
			low = mid + 1
		} else if errors.Is(err, domain.ErrNotFound) {
			high = mid - 1
		} else {
			return domain.Result{}, err
		}
	}

	if lastFilm.Id == 0 {
		return domain.Result{}, fmt.Errorf("Get last movie: %w", domain.ErrNotFound)
	}

	return lastFilm, nil
}
