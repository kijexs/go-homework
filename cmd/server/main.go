package main

import (
	"context"
	"fmt"
	"go-homework/internal/adapter"
	"go-homework/internal/adapter/postgres"
	"go-homework/internal/handler/grpcserver"
	"go-homework/internal/handler/httpserver"
	"go-homework/internal/migrations"
	"go-homework/internal/usecase"
	"go-homework/proto/pb"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go-homework/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// @title Movie API
// @version 1.0
// @description HTTP API for fetching movies from local PostgreSQL database
// @host localhost:8080
// @BasePath /
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	baseURL := os.Getenv("MOVIE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://homeworksite.site"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("create pgxpool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to postgres: %v", err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %v", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("apply migrations: %v", err)
	}

	syncWorkerCount := 50
	apiWorkerCount := 5

	providerDB := postgres.NewMovieRepository(pool)

	provider := adapter.NewMovieProvider(httpClient, baseURL, 5*time.Second)

	movieSyncService := usecase.NewMovieSyncService(providerDB, provider, syncWorkerCount)

	if err := movieSyncService.Sync(ctx); err != nil {
		return fmt.Errorf("failed to synchronize: %v", err)
	}

	movieService := usecase.NewMovieService(providerDB, apiWorkerCount)

	grpcHandler := grpcserver.NewGRPCHandler(movieService)

	listenerGRPC, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %v", err)
	}
	defer listenerGRPC.Close()

	grpcServer := grpc.NewServer()

	pb.RegisterMoviesServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(listenerGRPC); err != nil {
			fmt.Fprintf(os.Stderr, "gRPC server error: %v", err)
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
		return fmt.Errorf("failed to start HTTP server: %v", err)
	}

	return nil
}
