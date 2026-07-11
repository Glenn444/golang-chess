package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	Ping(ctx context.Context) error
}

type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

// NewStore creates a new store
func NewStore(dbPool *pgxpool.Pool) Store {
	return &SQLStore{
		db:      dbPool,
		Queries: New(dbPool),
	}
}

func (s *SQLStore) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}