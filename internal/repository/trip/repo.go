package trip

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

var _ Repository = (*Repo)(nil)

func New(pool *pgxpool.Pool, queryTimeout time.Duration) *Repo {
	return &Repo{pool: pool, queryTimeout: queryTimeout}
}
