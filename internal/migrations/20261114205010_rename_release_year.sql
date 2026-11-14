-- +goose Up
SELECT 'up SQL query';
ALTER TABLE movies
RENAME COLUMN release_year TO year;

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE movies
RENAME COLUMN year TO release_year;
