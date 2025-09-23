// pkg/database/database.go
package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"document-embeddings/internal/config"
)

type DB struct {
	*pgxpool.Pool
}

func New(cfg config.DatabaseConfig) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

