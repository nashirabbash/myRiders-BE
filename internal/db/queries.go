package db

import (
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queries defines the database query interface for handlers and services.
// In Phase 2, this is satisfied by code generated from sqlc.
type Queries = dbsqlc.Queries

// NewQueries creates a new Queries instance from a database connection pool
func NewQueries(pool *pgxpool.Pool) *Queries {
	return dbsqlc.New(pool)
}
