package repository

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	pool  *pgxpool.Pool
	cache Cache
}

func New(pool *pgxpool.Pool, cache Cache) *Repository {
	return &Repository{pool: pool, cache: cache}
}
