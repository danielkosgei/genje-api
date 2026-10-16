package repository

import "github.com/jackc/pgx/v5/pgxpool"

// SearchRepo matches names across the directory.
type SearchRepo struct {
	pool *pgxpool.Pool
}

func NewSearchRepo(pool *pgxpool.Pool) *SearchRepo {
	return &SearchRepo{pool: pool}
}

// RaceRepo reads county races.
type RaceRepo struct {
	pool *pgxpool.Pool
}

func NewRaceRepo(pool *pgxpool.Pool) *RaceRepo {
	return &RaceRepo{pool: pool}
}
