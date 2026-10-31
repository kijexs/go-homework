package main

import (
	"fmt"
	"go-homework/grpcserver"
	"go-homework/httpserver"
	"go-homework/movie"
	"go-homework/proto/pb"
	"net"
	"net/http"
	"os"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "go-homework/docs"                                                                                                                                                                                                                                                                                                                                        
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// @title Movie API
// @version 1.0                                                                                                                                                               
// @description HTTP API for fetching movies from an external movie service
// @host localhost:8080
// @BasePath /
func main() {
	baseURL := os.Getenv("MOVIE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://homeworksite.site"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	movieService := movie.NewMovieService(httpClient, baseURL, 5, 5*time.Second)

	grpcHandler := grpcserver.NewGRPCHandler(movieService)

	listenerGRPC, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start gRPC server: %v\n", err)
		os.Exit(1)
	}
	defer listenerGRPC.Close()

	grpcServer := grpc.NewServer()

	pb.RegisterMoviesServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(listenerGRPC); err != nil {
			fmt.Fprintf(os.Stderr, "gRPC server error: %v\n", err)
		}
	}()

	httpHandler := httpserver.NewHttpHadler(movieService)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /load-movies", httpHandler.HandleLoadMovies)
	mux.HandleFunc("GET /last-movie", httpHandler.HandleGetLastMovie)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start HTTP server: %v\n", err)
		os.Exit(1)
	}
}
