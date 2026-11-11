package domain

import "errors"

var (
	ErrInvalidArgument = errors.New("get invalid argument")
	ErrNotFound        = errors.New("movie not found")
	ErrUnavailable     = errors.New("movie service unavailable")
)
