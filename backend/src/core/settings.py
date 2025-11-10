"""
Centralized environment settings for all backend modules.
"""

from pathlib import Path

from pydantic import Field, computed_field, model_validator
from pydantic_settings import BaseSettings, SettingsConfigDict

from schema import JWTAlgorithm


_ENV_FILE = Path(__file__).resolve().parent / ".env"
_ENV_FILE_STR = str(_ENV_FILE) if _ENV_FILE.exists() else None


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=_ENV_FILE_STR, extra="allow")

    # PostgreSQL connection
    postgres_user: str = Field(..., description="Username used to connect to PostgreSQL")
    postgres_db: str = Field(..., description="PostgreSQL database name")

    postgres_port: int = Field(default=5432, description="PostgreSQL port")
    postgres_password: str = Field(
        default="",
        description=(
            "Password used to connect to PostgreSQL when a UNIX domain socket directory "
            "(POSTGRES_SOCKET_DIR) is not available."
        ),
    )
    postgres_host: str = Field(default="", description="PostgreSQL remote hostname or IP")
    postgres_socket_dir: str = Field(
        default="", description="PostgreSQL socket directory (preferred)"
    )
    postgres_pool_min_size: int = Field(
        default=1, description="Minimum number of connections in the PostgreSQL connection pool"
    )
    postgres_pool_max_size: int = Field(
        default=10, description="Maximum number of connections in the PostgreSQL connection pool"
    )
    postgres_max_queries: int = Field(
        default=50_000,
        description=(
            "Maximum number of queries executed by a single connection before it is closed "
            "and replaced with a new connection"
        ),
    )
    postgres_max_inactive_connection_lifetime: float = Field(
        default=300.0,
        description=(
            "Maximum lifetime (in seconds) for which a connection may remain idle in the pool "
            "before being closed"
        ),
    )
    postgres_command_timeout: float = Field(
        default=15.0,
        description="Default command timeout in seconds for asyncpg connections",
    )
    postgres_statement_cache_size: int = Field(
        default=200,
        description="Statement cache size for asyncpg connections",
    )
    # App runtime
    arch_stats_dev_mode: bool = Field(
        default=False, description="Enable development mode for Archy Stats"
    )
    arch_stats_server_port: int = Field(default=8000, description="Port for the Archy Stats server")
    arch_stats_ws_channel: str = Field(
        default="archy", description="WebSocket channel for Archy Stats"
    )
    apply_db_migrations_on_start: bool = Field(
        default=True, description="Apply database migrations automatically at startup"
    )

    # Auth settings (session cookie entropy / TTL only; external OAuth removed)
    session_ttl_hours: int = Field(default=24, description="Session lifetime in hours")
    session_token_bytes: int = Field(
        default=32, description="Number of random bytes (pre-encoding) for session token entropy"
    )

    # Google / JWT auth
    arch_stats_google_oauth_client_id: str = Field(
        ..., description="Google OAuth Client ID for One Tap / ID token audience validation"
    )
    arch_stats_jwt_secret: str = Field(
        default=...,
        description="HMAC secret used to sign JWTs (HS256). Provide strong secret in production.",
    )
    arch_stats_jwt_algorithm: JWTAlgorithm = Field(
        default=JWTAlgorithm.HS256, description="JWT signing algorithm"
    )
    arch_stats_jwt_ttl_minutes: int = Field(
        default=60, description="JWT access token lifetime (minutes)"
    )

    @computed_field  # type: ignore[prop-decorator]
    @property
    def postgres_dsn_host(self) -> str:
        """Computed field: returns socket directory or TCP host based on what's available.

        Returns socket_dir if it exists and contains the PostgreSQL socket file,
        otherwise returns the TCP host. This allows the rest of the app to use
        a single field regardless of the connection method.

        Returns:
            The PostgreSQL connection host (socket directory path or hostname/IP).

        Raises:
            ValueError: If neither valid socket nor TCP host is available.
        """
        # Prefer Unix socket if the socket file exists
        if self.postgres_socket_dir:
            socket_dir = Path(self.postgres_socket_dir)
            socket_file = socket_dir / f".s.PGSQL.{self.postgres_port}"
            if socket_dir.is_dir() and socket_file.exists():
                return self.postgres_socket_dir

        # Fallback to TCP host if set
        if self.postgres_host:
            return self.postgres_host

        raise ValueError(
            "Database connection requires either an active Unix socket at "
            f"'{self.postgres_socket_dir}' or a TCP host via 'postgres_host'."
        )

    @model_validator(mode="after")
    def validate_postgres_connection(self) -> "Settings":
        """Validate PostgreSQL connection configuration.

        In production (arch_stats_dev_mode=False), postgres_socket_dir MUST be provided
        and point to a valid directory. In development mode, we allow fallback to
        host/port/password if socket directory is not available.

        Raises:
            ValueError: If production mode is enabled but postgres_socket_dir is invalid.
        """
        if not self.arch_stats_dev_mode:
            # Production mode: require valid socket directory
            if not self.postgres_socket_dir:
                raise ValueError(
                    "Production mode (arch_stats_dev_mode=False) requires POSTGRES_SOCKET_DIR "
                    "to be set. Socket-based authentication is mandatory for production."
                )

            socket_path = Path(self.postgres_socket_dir)
            if not socket_path.exists():
                raise ValueError(
                    f"Production mode requires a valid POSTGRES_SOCKET_DIR. "
                    f"Directory does not exist: {self.postgres_socket_dir}"
                )

            if not socket_path.is_dir():
                raise ValueError(
                    f"Production mode requires POSTGRES_SOCKET_DIR to be a directory. "
                    f"Path is not a directory: {self.postgres_socket_dir}"
                )

        return self


settings = Settings()
