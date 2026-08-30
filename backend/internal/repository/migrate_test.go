package repository_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestMigrationFilesExistAndHaveGooseMarkers(t *testing.T) {
	migrationsDir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("reading migrations dir: %v", err)
	}

	pattern := regexp.MustCompile(`^\d{3}_[\w-]+\.sql$`)
	sqlCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		sqlCount++
		name := entry.Name()
		if !pattern.MatchString(name) {
			t.Errorf("migration file %q does not match pattern NNN_name.sql", name)
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Errorf("reading %s: %v", name, err)
			continue
		}
		strContent := string(content)
		if !strings.Contains(strContent, "-- +goose Up") {
			t.Errorf("migration file %s missing '-- +goose Up' marker", name)
		}
		if !strings.Contains(strContent, "-- +goose Down") {
			t.Errorf("migration file %s missing '-- +goose Down' marker", name)
		}
	}

	if sqlCount == 0 {
		t.Fatal("no SQL migration files found in migrations directory")
	}
}
