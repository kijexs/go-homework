package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"go-homework/internal/domain"
	"net/http"
	"strconv"
)

type httpHandler struct {
	movies domain.MovieUseCase
}

func NewHttpHadler(movies domain.MovieUseCase) *httpHandler {
	return &httpHandler{movies: movies}
}

type httpMovieResponse struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int64  `json:"year"`
	Director string `json:"director"`
	Err      string `json:"error,omitempty"`
}

func toHttpMovieResponse(res domain.Result) httpMovieResponse {
	errMessage := ""
	if res.Err != nil {
		errMessage = res.Err.Error()
	}

	return httpMovieResponse{
		Id:       int64(res.Id),
		Title:    res.Data.Title,
		Year:     int64(res.Data.Year),
		Director: res.Data.Director,
		Err:      errMessage,
	}
}

func mapHTTPError(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, domain.ErrInvalidArgument):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// LoadMovies godoc
// @Summary Get movies information
// @Description Returns list of movies
// @Tags load-movies
// @Produce json
// @Param from query int true "Movie from"
// @Param to query int true "Movie to"
// @Success 200 {array} httpMovieResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Failure 503 {string} string
// @Failure 504 {string} string
// @Router /load-movies [get]
func (s *httpHandler) HandleLoadMovies(w http.ResponseWriter, r *http.Request) {
	from, err := strconv.Atoi(r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	movieRequest := domain.MovieRequest{From: from, To: to}
	res, err := s.movies.LoadMovies(r.Context(), movieRequest)
	if err != nil {
		http.Error(w, err.Error(), mapHTTPError(err))
		return
	}

	httpMovieRes := make([]httpMovieResponse, 0, len(res))
	for _, mov := range res {
		httpMovieRes = append(httpMovieRes, toHttpMovieResponse(mov))
	}

	movieJSON, err := json.Marshal(httpMovieRes)
	if err != nil {
		http.Error(w, err.Error(), mapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(movieJSON)
}

// GetLastMovie godoc
// @Summary Get last movies information
// @Description Returns information about last movie
// @Tags last-movie
// @Produce json
// @Success 200 {object} httpMovieResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Failure 503 {string} string
// @Failure 504 {string} string
// @Router /last-movie [get]
func (s *httpHandler) HandleGetLastMovie(w http.ResponseWriter, r *http.Request) {
	res, err := s.movies.GetLastMovie(r.Context())
	if err != nil {
		http.Error(w, err.Error(), mapHTTPError(err))
		return
	}

	movieJSON, err := json.Marshal(toHttpMovieResponse(res))
	if err != nil {
		http.Error(w, err.Error(), mapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(movieJSON)
}
