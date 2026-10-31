package grpcserver

import (
	"context"
	"errors"
	"go-homework/movie"
	"go-homework/proto/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type grpcHandler struct {
	pb.UnimplementedMoviesServiceServer
	movies *movie.MovieService
}

func NewGRPCHandler(movies *movie.MovieService) *grpcHandler {
	return &grpcHandler{movies: movies}
}

func mapGRPCError(err error) error {
	switch {
	case errors.Is(err, movie.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, movie.ErrUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, movie.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (c *grpcHandler) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	request := movie.MovieRequest{To: int(req.To), From: int(req.From)}
	res, err := c.movies.LoadMovies(ctx, request)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	moviesList := []*pb.MovieResponse{}
	for _, mov := range *res {
		message := ""
		if mov.Err != nil {
			message = mov.Err.Error()
		}

		movieResponse := pb.MovieResponse{
			Id:       int64(mov.Id),
			Title:    mov.Data.Title,
			Year:     int64(mov.Data.Year),
			Director: mov.Data.Director,
			Message:  message,
		}
		moviesList = append(moviesList, &movieResponse)
	}
	return &pb.LoadMoviesResponse{Movies: moviesList}, nil
}

func (c *grpcHandler) GetLastMovie(ctx context.Context, empty *emptypb.Empty) (*pb.MovieResponse, error) {
	res, err := c.movies.GetLastMovie(ctx)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	message := ""
	if res.Err != nil {
		message = res.Err.Error()
	}

	movieResponse := pb.MovieResponse{
		Id:       int64(res.Id),
		Title:    res.Data.Title,
		Year:     int64(res.Data.Year),
		Director: res.Data.Director,
		Message:  message,
	}
	return &movieResponse, nil
}
