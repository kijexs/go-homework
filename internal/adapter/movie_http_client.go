package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-homework/internal/domain"
)

type Provider struct {
	httpClient   *http.Client
	baseURL      string
	fetchTimeout time.Duration
}

func NewMovieProvider(httpClient *http.Client, baseURL string, fetchTimeout time.Duration) *Provider {
	if fetchTimeout <= 0 {
		fetchTimeout = 5 * time.Second
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &Provider{
		httpClient:   httpClient,
		baseURL:      baseURL,
		fetchTimeout: fetchTimeout,
	}
}

type jsonMovie struct {
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func (p *Provider) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, p.fetchTimeout)
	defer cancel()

	var req, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		p.baseURL+"/"+strconv.Itoa(id)+"/info.0.json",
		nil,
	)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("error forming a new request: %w", err)
	}
	res, err := p.httpClient.Do(req)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return domain.Movie{}, fmt.Errorf("movie %d: request canceled: %w", id, err)

		case errors.Is(err, context.DeadlineExceeded):
			return domain.Movie{}, fmt.Errorf("movie %d: request timeout: %w", id, err)

		default:
			return domain.Movie{}, fmt.Errorf("movie %d: %w: %v", id, domain.ErrUnavailable, err)
		}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch {
		case res.StatusCode == http.StatusNotFound:
			return domain.Movie{}, fmt.Errorf("movie %d: %w", id, domain.ErrNotFound)

		case res.StatusCode >= 500:
			return domain.Movie{}, fmt.Errorf("movie %d: %w", id, domain.ErrUnavailable)

		default:
			return domain.Movie{}, fmt.Errorf("movie %d: unexpected status %d", id, res.StatusCode)
		}
	}

	var movie jsonMovie
	err = json.NewDecoder(res.Body).Decode(&movie)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("decode response: %w", err)
	}

	return domain.Movie{
		Title:    movie.Title,
		Year:     movie.Year,
		Director: movie.Director,
	}, nil
}
