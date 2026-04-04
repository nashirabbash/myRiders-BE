package db

import "github.com/jackc/pgx/v5/pgxpool"

// Store provides access to database queries.
// This type wraps sqlc.Queries and will be replaced by a properly generated
// store interface from sqlc in Phase 2 when the schema is complete.
type Store struct {
	pool *pgxpool.Pool
	// Queries will be embedded here: *sqlc.Queries
}

// NewStore creates a new Store instance.
// In Phase 2, this will initialize the actual sqlc.Queries with generated methods.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{pool: db}
}
