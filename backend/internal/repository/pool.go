// Package repository provides database access functions.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ParsePoolConfig parses a DSN and returns a pgxpool.Config with the
// specified min/max connection counts.
func ParsePoolConfig(dsn string, minConns, maxConns int32) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing pool config: %w", err)
	}
	cfg.MinConns = minConns
	cfg.MaxConns = maxConns
	return cfg, nil
}

// NewPool creates a new pgxpool.Pool from the given DSN and pool size parameters.
// The caller is responsible for calling pool.Close() when done.
func NewPool(ctx context.Context, dsn string, minConns, maxConns int32) (*pgxpool.Pool, error) {
	poolCfg, err := ParsePoolConfig(dsn, minConns, maxConns)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
