package sqlc

import "github.com/jackc/pgx/v5/pgxpool"

// Queries provides database query methods.
// This type will be extended by sqlc code generation in Phase 2.
type Queries struct {
	db *pgxpool.Pool
}

// New creates a new Queries instance.
func New(db *pgxpool.Pool) *Queries {
	return &Queries{db: db}
}
