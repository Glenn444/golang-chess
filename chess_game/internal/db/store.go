package db

import (

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface{
	Querier
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