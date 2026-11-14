-- +goose Up
SELECT 'up SQL query';
CREATE TABLE movies (
  id bigint PRIMARY KEY,

  title text NOT NULL,

  release_year integer NOT NULL,

  director text NOT NULL
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE movies;