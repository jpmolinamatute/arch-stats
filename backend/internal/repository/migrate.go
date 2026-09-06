package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations applies pending database migrations from the given directory.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("running goose up migrations: %w", err)
	}

	return nil
}

// RollbackMigration rolls back the last applied migration.
func RollbackMigration(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("running goose down migration: %w", err)
	}

	return nil
}

// GetSchemaVersion returns the current database schema version.
func GetSchemaVersion(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("setting goose dialect: %w", err)
	}

	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("getting db version: %w", err)
	}

	return version, nil
}
