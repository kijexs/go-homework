package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-homework/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository struct {
	pool *pgxpool.Pool
}

func NewMovieRepository(pool *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{
		pool: pool,
	}
}

func (r *MovieRepository) GetByID(
	ctx context.Context, id int,
) (domain.Movie, error) {
	var m domain.Movie
	err := r.pool.QueryRow(ctx,
		`SELECT title, release_year, director
         FROM movies WHERE id = $1`, id,
	).Scan(&m.Title, &m.Year, &m.Director)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Movie{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Movie{}, fmt.Errorf("get movie %d: %w", id, err)
	}

	return m, nil
}

func (r *MovieRepository) Save(
	ctx context.Context, id int, movie domain.Movie,
) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO movies (id, title, release_year, director)
		VALUES($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING;`,
		id, movie.Title, movie.Year, movie.Director,
	)
	if err != nil {
		return fmt.Errorf("save movie %d: %w", id, err)
	}
	return nil
}

func (r *MovieRepository) GetLast(
	ctx context.Context,
) (domain.Result, error) {
	var m domain.Result
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, release_year, director
		FROM movies
		ORDER BY id DESC
		LIMIT 1`,
	).Scan(&m.Id, &m.Data.Title, &m.Data.Year, &m.Data.Director)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Result{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Result{}, fmt.Errorf("get last movie: %w", err)
	}
	return m, nil
}

func (r *MovieRepository) GetNotExistingIDs(
	ctx context.Context, maxId int,
) ([]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT candidate.id
		FROM generate_series(1, $1) AS candidate(id)
		WHERE NOT EXISTS (
			SELECT 1
			FROM movies AS m
			WHERE m.id = candidate.id
		)
		ORDER BY candidate.id;`, maxId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return []int{}, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
