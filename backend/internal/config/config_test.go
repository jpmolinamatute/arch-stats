package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
)

func setAllRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_POOL_MIN_SIZE", "1")
	t.Setenv("POSTGRES_POOL_MAX_SIZE", "10")
	t.Setenv("POSTGRES_MAX_QUERIES", "50000")
	t.Setenv("POSTGRES_MAX_INACTIVE_CONNECTION_LIFETIME", "300.0")
	t.Setenv("POSTGRES_COMMAND_TIMEOUT", "15.0")
	t.Setenv("POSTGRES_STATEMENT_CACHE_SIZE", "200")
	t.Setenv("ARCH_STATS_DEV_MODE", "true")
	t.Setenv("ARCH_STATS_SERVER_PORT", "8000")
	t.Setenv("ARCH_STATS_WS_CHANNEL", "archy")
	t.Setenv("APPLY_DB_MIGRATIONS_ON_START", "true")
	t.Setenv("SESSION_TTL_HOURS", "24")
	t.Setenv("SESSION_TOKEN_BYTES", "32")
	t.Setenv("ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID", "test-client-id")
	t.Setenv("ARCH_STATS_JWT_SECRET", "test-secret-key-minimum-length")
	t.Setenv("ARCH_STATS_JWT_ALGORITHM", "HS256")
	t.Setenv("ARCH_STATS_JWT_TTL_MINUTES", "60")
}

func TestLoadConfig(t *testing.T) {
	setAllRequiredEnv(t)
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("ARCH_STATS_SERVER_PORT", "9000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.PostgresUser != "testuser" {
		t.Errorf("PostgresUser = %q, want %q", cfg.PostgresUser, "testuser")
	}
	if cfg.PostgresPassword != "testpass" {
		t.Errorf("PostgresPassword = %q, want %q", cfg.PostgresPassword, "testpass")
	}
	if cfg.PostgresDB != "testdb" {
		t.Errorf("PostgresDB = %q, want %q", cfg.PostgresDB, "testdb")
	}
	if cfg.PostgresPort != 5433 {
		t.Errorf("PostgresPort = %d, want %d", cfg.PostgresPort, 5433)
	}
	if cfg.PostgresPoolMinSize != 1 {
		t.Errorf("PostgresPoolMinSize = %d, want 1", cfg.PostgresPoolMinSize)
	}
	if cfg.PostgresPoolMaxSize != 10 {
		t.Errorf("PostgresPoolMaxSize = %d, want 10", cfg.PostgresPoolMaxSize)
	}
	if cfg.PostgresMaxQueries != 50000 {
		t.Errorf("PostgresMaxQueries = %d, want 50000", cfg.PostgresMaxQueries)
	}
	if cfg.PostgresMaxInactiveConnectionLifetime != 300.0 {
		t.Errorf("PostgresMaxInactiveConnectionLifetime = %f, want 300.0", cfg.PostgresMaxInactiveConnectionLifetime)
	}
	if cfg.PostgresCommandTimeout != 15.0 {
		t.Errorf("PostgresCommandTimeout = %f, want 15.0", cfg.PostgresCommandTimeout)
	}
	if cfg.PostgresStatementCacheSize != 200 {
		t.Errorf("PostgresStatementCacheSize = %d, want 200", cfg.PostgresStatementCacheSize)
	}
	if !cfg.DevMode {
		t.Errorf("DevMode = %v, want true", cfg.DevMode)
	}
	if cfg.ServerPort != 9000 {
		t.Errorf("ServerPort = %d, want %d", cfg.ServerPort, 9000)
	}
	if cfg.WSChannel != "archy" {
		t.Errorf("WSChannel = %q, want archy", cfg.WSChannel)
	}
	if !cfg.ApplyMigrationsOnStart {
		t.Errorf("ApplyMigrationsOnStart = %v, want true", cfg.ApplyMigrationsOnStart)
	}
	if cfg.SessionTTLHours != 24 {
		t.Errorf("SessionTTLHours = %d, want 24", cfg.SessionTTLHours)
	}
	if cfg.SessionTokenBytes != 32 {
		t.Errorf("SessionTokenBytes = %d, want 32", cfg.SessionTokenBytes)
	}
	if cfg.GoogleOAuthClientID != "test-client-id" {
		t.Errorf("GoogleOAuthClientID = %q, want test-client-id", cfg.GoogleOAuthClientID)
	}
	if cfg.JWTSecret != "test-secret-key-minimum-length" {
		t.Errorf("JWTSecret = %q, want test-secret-key-minimum-length", cfg.JWTSecret)
	}
	if cfg.JWTAlgorithm != "HS256" {
		t.Errorf("JWTAlgorithm = %q, want HS256", cfg.JWTAlgorithm)
	}
	if cfg.JWTTTLMinutes != 60 {
		t.Errorf("JWTTTLMinutes = %d, want 60", cfg.JWTTTLMinutes)
	}
}

func TestLoadConfig_MissingRequiredVarFails(t *testing.T) {
	requiredVars := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"POSTGRES_PORT",
		"POSTGRES_POOL_MIN_SIZE",
		"POSTGRES_POOL_MAX_SIZE",
		"POSTGRES_MAX_QUERIES",
		"POSTGRES_MAX_INACTIVE_CONNECTION_LIFETIME",
		"POSTGRES_COMMAND_TIMEOUT",
		"POSTGRES_STATEMENT_CACHE_SIZE",
		"ARCH_STATS_DEV_MODE",
		"ARCH_STATS_SERVER_PORT",
		"ARCH_STATS_WS_CHANNEL",
		"APPLY_DB_MIGRATIONS_ON_START",
		"SESSION_TTL_HOURS",
		"SESSION_TOKEN_BYTES",
		"ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID",
		"ARCH_STATS_JWT_SECRET",
		"ARCH_STATS_JWT_ALGORITHM",
		"ARCH_STATS_JWT_TTL_MINUTES",
	}

	for _, v := range requiredVars {
		t.Run("missing_"+v, func(t *testing.T) {
			setAllRequiredEnv(t)
			os.Unsetenv(v)
			_, err := config.Load()
			if err == nil {
				t.Errorf("Load() should fail when %s is undefined", v)
			}
		})
	}
}

func TestDSNHost_SocketFallbackToTCP(t *testing.T) {
	setAllRequiredEnv(t)
	t.Setenv("POSTGRES_HOST", "db.example.com")
	os.Unsetenv("POSTGRES_SOCKET_DIR")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	host, err := cfg.DSNHost()
	if err != nil {
		t.Fatalf("DSNHost() error: %v", err)
	}
	if host != "db.example.com" {
		t.Errorf("DSNHost() = %q, want %q", host, "db.example.com")
	}
}

func TestDSNHost_SocketDirWithSocketFile(t *testing.T) {
	setAllRequiredEnv(t)
	tmpDir := t.TempDir()
	socketFile := filepath.Join(tmpDir, ".s.PGSQL.5432")
	if err := os.WriteFile(socketFile, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create socket file: %v", err)
	}

	t.Setenv("POSTGRES_SOCKET_DIR", tmpDir)
	t.Setenv("POSTGRES_HOST", "db.example.com")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	host, err := cfg.DSNHost()
	if err != nil {
		t.Fatalf("DSNHost() error: %v", err)
	}
	if host != tmpDir {
		t.Errorf("DSNHost() = %q, want %q", host, tmpDir)
	}
}

func TestDSNHost_NeitherSetReturnsError(t *testing.T) {
	setAllRequiredEnv(t)
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("POSTGRES_SOCKET_DIR")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	_, err = cfg.DSNHost()
	if err == nil {
		t.Error("DSNHost() should return error when neither host nor socket is set")
	}
}

func TestValidate_ProductionRequiresSocketDir(t *testing.T) {
	setAllRequiredEnv(t)
	t.Setenv("ARCH_STATS_DEV_MODE", "false")
	os.Unsetenv("POSTGRES_SOCKET_DIR")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("Validate() should fail in production without POSTGRES_SOCKET_DIR")
	}
}

func TestValidate_ProductionWithNonExistentSocketDir(t *testing.T) {
	setAllRequiredEnv(t)
	t.Setenv("ARCH_STATS_DEV_MODE", "false")
	t.Setenv("POSTGRES_SOCKET_DIR", "/non/existent/path/for/socket")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("Validate() should fail in production when POSTGRES_SOCKET_DIR does not exist")
	}
}

func TestValidate_ProductionWithValidSocketDir(t *testing.T) {
	setAllRequiredEnv(t)
	tmpDir := t.TempDir()
	t.Setenv("ARCH_STATS_DEV_MODE", "false")
	t.Setenv("POSTGRES_SOCKET_DIR", tmpDir)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	err = cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should pass in production with existing socket dir: %v", err)
	}
}

func TestValidate_DevModeAllowsTCP(t *testing.T) {
	setAllRequiredEnv(t)
	t.Setenv("ARCH_STATS_DEV_MODE", "true")
	t.Setenv("POSTGRES_HOST", "localhost")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	err = cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should pass in dev mode with TCP host: %v", err)
	}
}

func TestDatabaseURL_TCP(t *testing.T) {
	setAllRequiredEnv(t)
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "user1")
	t.Setenv("POSTGRES_PASSWORD", "pass1")
	t.Setenv("POSTGRES_DB", "db1")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	dbURL, err := cfg.DatabaseURL()
	if err != nil {
		t.Fatalf("DatabaseURL() error: %v", err)
	}

	expected := "postgres://user1:pass1@localhost:5432/db1"
	if dbURL != expected {
		t.Errorf("DatabaseURL() = %q, want %q", dbURL, expected)
	}
}

func TestDatabaseURL_UnixSocket(t *testing.T) {
	setAllRequiredEnv(t)
	tmpDir := t.TempDir()
	socketFile := filepath.Join(tmpDir, ".s.PGSQL.5432")
	if err := os.WriteFile(socketFile, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create socket file: %v", err)
	}

	t.Setenv("POSTGRES_SOCKET_DIR", tmpDir)
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "user1")
	t.Setenv("POSTGRES_PASSWORD", "pass1")
	t.Setenv("POSTGRES_DB", "db1")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	dbURL, err := cfg.DatabaseURL()
	if err != nil {
		t.Fatalf("DatabaseURL() error: %v", err)
	}

	if !strings.HasPrefix(dbURL, "postgres://user1:pass1@/db1?") || !strings.Contains(dbURL, "host=") {
		t.Errorf("DatabaseURL() for socket = %q, expected socket URL format with host param", dbURL)
	}
}
