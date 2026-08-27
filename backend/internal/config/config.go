// Package config manages application configuration loaded from environment variables.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration settings for the backend application.
type Config struct {
	// PostgreSQL connection settings
	PostgresUser                          string  `envconfig:"POSTGRES_USER" required:"true"`
	PostgresPassword                      string  `envconfig:"POSTGRES_PASSWORD" required:"true"`
	PostgresDB                            string  `envconfig:"POSTGRES_DB" required:"true"`
	PostgresHost                          string  `envconfig:"POSTGRES_HOST"`
	PostgresPort                          int     `envconfig:"POSTGRES_PORT" required:"true"`
	PostgresSocketDir                     string  `envconfig:"POSTGRES_SOCKET_DIR"`
	PostgresPoolMinSize                   int     `envconfig:"POSTGRES_POOL_MIN_SIZE" required:"true"`
	PostgresPoolMaxSize                   int     `envconfig:"POSTGRES_POOL_MAX_SIZE" required:"true"`
	PostgresMaxQueries                    int     `envconfig:"POSTGRES_MAX_QUERIES" required:"true"`
	PostgresMaxInactiveConnectionLifetime float64 `envconfig:"POSTGRES_MAX_INACTIVE_CONNECTION_LIFETIME" required:"true"`
	PostgresCommandTimeout                float64 `envconfig:"POSTGRES_COMMAND_TIMEOUT" required:"true"`
	PostgresStatementCacheSize            int     `envconfig:"POSTGRES_STATEMENT_CACHE_SIZE" required:"true"`

	// App runtime settings
	DevMode                bool   `envconfig:"ARCH_STATS_DEV_MODE" required:"true"`
	ServerPort             int    `envconfig:"ARCH_STATS_SERVER_PORT" required:"true"`
	WSChannel              string `envconfig:"ARCH_STATS_WS_CHANNEL" required:"true"`
	ApplyMigrationsOnStart bool   `envconfig:"APPLY_DB_MIGRATIONS_ON_START" required:"true"`

	// Session settings
	SessionTTLHours   int `envconfig:"SESSION_TTL_HOURS" required:"true"`
	SessionTokenBytes int `envconfig:"SESSION_TOKEN_BYTES" required:"true"`

	// Google / JWT auth settings
	GoogleOAuthClientID string `envconfig:"ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID" required:"true"`
	JWTSecret           string `envconfig:"ARCH_STATS_JWT_SECRET" required:"true"`
	JWTAlgorithm        string `envconfig:"ARCH_STATS_JWT_ALGORITHM" required:"true"`
	JWTTTLMinutes       int    `envconfig:"ARCH_STATS_JWT_TTL_MINUTES" required:"true"`
}

// Load loads application configuration from environment variables and an optional .env file.
// It fails if any required environment variable is undefined.
func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	return &cfg, nil
}

// DSNHost returns the socket directory if an active Unix socket exists at that directory,
// or falls back to the TCP host.
func (c *Config) DSNHost() (string, error) {
	if c.PostgresSocketDir != "" {
		socketFile := filepath.Join(c.PostgresSocketDir, fmt.Sprintf(".s.PGSQL.%d", c.PostgresPort))
		if info, err := os.Stat(socketFile); err == nil && !info.IsDir() {
			return c.PostgresSocketDir, nil
		}
	}

	if c.PostgresHost != "" {
		return c.PostgresHost, nil
	}

	return "", fmt.Errorf(
		"database connection requires either an active Unix socket at %q or a TCP host via POSTGRES_HOST",
		c.PostgresSocketDir,
	)
}

// Validate performs cross-field validation on the configuration.
// In production mode (DevMode=false), PostgresSocketDir must exist and be a valid directory.
// In development mode, either PostgresSocketDir or PostgresHost must be provided.
func (c *Config) Validate() error {
	if !c.DevMode {
		if c.PostgresSocketDir == "" {
			return fmt.Errorf("production mode (ARCH_STATS_DEV_MODE=false) requires POSTGRES_SOCKET_DIR to be set")
		}

		info, err := os.Stat(c.PostgresSocketDir)
		if err != nil {
			return fmt.Errorf("production mode requires a valid POSTGRES_SOCKET_DIR: directory does not exist: %s", c.PostgresSocketDir)
		}

		if !info.IsDir() {
			return fmt.Errorf("production mode requires POSTGRES_SOCKET_DIR to be a directory: %s", c.PostgresSocketDir)
		}

		return nil
	}

	if c.PostgresSocketDir == "" && c.PostgresHost == "" {
		return fmt.Errorf("development mode requires either POSTGRES_SOCKET_DIR or POSTGRES_HOST to be set")
	}

	return nil
}

// DatabaseURL constructs the PostgreSQL connection string.
func (c *Config) DatabaseURL() (string, error) {
	host, err := c.DSNHost()
	if err != nil {
		return "", err
	}

	u := &url.URL{
		Scheme: "postgres",
		Path:   "/" + c.PostgresDB,
	}

	if c.PostgresPassword != "" {
		u.User = url.UserPassword(c.PostgresUser, c.PostgresPassword)
	} else if c.PostgresUser != "" {
		u.User = url.User(c.PostgresUser)
	}

	if strings.HasPrefix(host, "/") {
		q := url.Values{}
		q.Set("host", host)
		if c.PostgresPort != 0 {
			q.Set("port", strconv.Itoa(c.PostgresPort))
		}
		u.RawQuery = q.Encode()
	} else {
		if c.PostgresPort != 0 {
			u.Host = net.JoinHostPort(host, strconv.Itoa(c.PostgresPort))
		} else {
			u.Host = host
		}
	}

	return u.String(), nil
}
