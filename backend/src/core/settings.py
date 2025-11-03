"""
Centralized environment settings for all backend modules.
"""

from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


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
    postgres_host: str | None = Field(default=None, description="PostgreSQL remote hostname or IP")
    postgres_socket_dir: str | None = Field(
        default=None, description="PostgreSQL socket directory (preferred)"
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
    google_oauth_client_id: str = Field(
        ..., description="Google OAuth Client ID for One Tap / ID token audience validation"
    )
    jwt_secret: str = Field(
        default="dev-insecure-change-me",
        description="HMAC secret used to sign JWTs (HS256). Provide strong secret in production.",
    )
    jwt_algorithm: str = Field(default="HS256", description="JWT signing algorithm")
    jwt_ttl_minutes: int = Field(default=60, description="JWT access token lifetime (minutes)")


settings = Settings()
