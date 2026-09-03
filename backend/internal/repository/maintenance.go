package repository

import (
	"context"
	"fmt"
)

// MaintenanceRepo manages administrative and maintenance database operations.
type MaintenanceRepo struct {
	db DBTX
}

// NewMaintenanceRepo constructs a new MaintenanceRepo backed by DBTX.
func NewMaintenanceRepo(db DBTX) *MaintenanceRepo {
	return &MaintenanceRepo{db: db}
}

// RefreshOpenParticipants concurrently refreshes the open_participants materialized view.
func (r *MaintenanceRepo) RefreshOpenParticipants(ctx context.Context) error {
	if _, err := r.db.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY open_participants"); err != nil {
		return fmt.Errorf("refreshing open_participants materialized view: %w", err)
	}
	return nil
}

// GetSchemaVersion retrieves the current applied goose schema version from goose_db_version.
func (r *MaintenanceRepo) GetSchemaVersion(ctx context.Context) (int64, error) {
	var version int64
	row := r.db.QueryRow(ctx, "SELECT COALESCE(MAX(version_id), 0) FROM goose_db_version WHERE is_applied = true")
	if err := row.Scan(&version); err != nil {
		return 0, fmt.Errorf("getting schema version: %w", err)
	}
	return version, nil
}
