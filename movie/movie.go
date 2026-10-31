package movie

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MovieService struct {
	httpClient   *http.Client
	baseURL      string
	workerCount  int
	fetchTimeout time.Duration
}

func NewMovieService(httpClient *http.Client, baseURL string, workerCount int, fetchTimeout time.Duration) *MovieService {
	if workerCount <= 0 {
		workerCount = 1
	}
	if fetchTimeout <= 0 {
		fetchTimeout = 5 * time.Second
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &MovieService{httpClient: httpClient, baseURL: baseURL, workerCount: workerCount, fetchTimeout: fetchTimeout}
}

type MovieRequest struct {
	To   int
	From int
}

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

var (
    ErrNotFound    = errors.New("movie not found")
    ErrUnavailable = errors.New("movie service unavailable")
	ErrInvalidArgument = errors.New("get invalid argument")
)

func fetchMovie(ctx context.Context, service MovieService, movieId int) Result {
	result := Result{Id: movieId}
	ctx, cancel := context.WithTimeout(ctx, service.fetchTimeout)
	defer cancel()

	var req, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		service.baseURL+"/"+strconv.Itoa(movieId)+"/info.0.json",
		nil,
	)
	if err != nil {
		result.Err = fmt.Errorf("error forming a new request: %w", err)
		return result
	}
	res, err := service.httpClient.Do(req)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			result.Err = fmt.Errorf("movie %d: request canceled: %w", movieId, err)

		case errors.Is(err, context.DeadlineExceeded):
			result.Err = fmt.Errorf("movie %d: request timeout: %w", movieId, err)

		default:
			result.Err = fmt.Errorf("movie %d: %w: %v", movieId, ErrUnavailable, err)
		}
		return result
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch {
		case res.StatusCode == http.StatusNotFound:
			result.Err = fmt.Errorf("movie %d: %w", movieId, ErrNotFound)

		case res.StatusCode >= 500:
			result.Err = fmt.Errorf("movie %d: %w", movieId, ErrUnavailable)

		default:
			result.Err = fmt.Errorf("movie %d: unexpected status %d", movieId, res.StatusCode)
		}
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

func (s *MovieService) LoadMovies(ctx context.Context, req MovieRequest) (*[]Result, error) {
	if req.From <= 0 || req.To <= 0 || req.From > req.To {
		return nil, fmt.Errorf("Load movies: %w", ErrInvalidArgument)
	}
	result := []Result{}
	var wg sync.WaitGroup

	jobs := make(chan int)
	results := make(chan Result)

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
	return &result, nil
}

func (s *MovieService) GetLastMovie(ctx context.Context) (*Result, error) {
	low := 1
	hight := 1
	
	for {
		res := fetchMovie(ctx, *s, hight)
		
		if errors.Is(res.Err, ErrNotFound) {
			break
		}
		if res.Err != nil {
			return  nil, res.Err
		}
		low = hight
		hight *= 2
	}

	res := Result{}
	lastFilm := Result{}

	for low <= hight {
		mid := low + ((hight - low) / 2)
		res = fetchMovie(ctx, *s, mid)
		if res.Err == nil {
			lastFilm = res
			low = mid + 1
		} else if errors.Is(res.Err, ErrNotFound) {
			hight = mid - 1
		} else {
			return nil, res.Err
		}
	}

	if lastFilm.Id == 0 {
		return nil, fmt.Errorf("Get last movie: %w", ErrNotFound)
	}
	
	return &lastFilm, nil
}
