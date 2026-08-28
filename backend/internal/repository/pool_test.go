package repository_test

import (
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestParsePoolConfig(t *testing.T) {
	dsn := "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	poolCfg, err := repository.ParsePoolConfig(dsn, 2, 10)
	if err != nil {
		t.Fatalf("ParsePoolConfig() error: %v", err)
	}

	if poolCfg.MinConns != 2 {
		t.Errorf("MinConns = %d, want 2", poolCfg.MinConns)
	}
	if poolCfg.MaxConns != 10 {
		t.Errorf("MaxConns = %d, want 10", poolCfg.MaxConns)
	}
}

func TestParsePoolConfig_InvalidDSN(t *testing.T) {
	_, err := repository.ParsePoolConfig("not-a-valid-dsn://", 1, 5)
	if err == nil {
		t.Error("ParsePoolConfig() should fail with invalid DSN")
	}
}
