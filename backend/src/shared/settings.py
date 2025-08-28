"""
Centralized environment settings for all backend modules.
"""

from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


# Resolve .env location relative to this file (shared-only)
_ENV_FILE = Path(__file__).resolve().parent / ".env"
_ENV_FILE_STR = str(_ENV_FILE) if _ENV_FILE.exists() else None


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=_ENV_FILE_STR)

    # PostgreSQL connection
    postgres_user: str = Field(..., description="Username used to connect to PostgreSQL")
    postgres_db: str = Field(..., description="PostgreSQL database name")

    postgres_port: int = Field(default=5432, description="PostgreSQL port")
    postgres_password: str = Field(
        default="",
        description=(
            "password used to connect to PostgreSQL in case not having POSTGRES_SOCKET_DIR"
        ),
    )
    postgres_host: str = Field(default="", description="PostgreSQL remote hostname or IP")
    postgres_socket_dir: str = Field(
        default="/var/run/postgresql", description="PostgreSQL socket directory (preferred)"
    )
    pg_pool_min: int = Field(
        default=1, description="Minimum number of connections in the PostgreSQL connection pool"
    )
    pg_pool_max: int = Field(
        default=10, description="Maximum number of connections in the PostgreSQL connection pool"
    )

    # App runtime
    arch_stats_dev_mode: bool = Field(
        default=False, description="Enable development mode for Archy Stats"
    )
    arch_stats_server_port: int = Field(default=8000, description="Port for the Archy Stats server")
    arch_stats_ws_channel: str = Field(
        default="archy", description="WebSocket channel for Archy Stats"
    )

    command_timeout: float | None = Field(
        default=15.0,
        description="Default command timeout in seconds for asyncpg connections",
    )
    statement_cache_size: int | None = Field(
        default=200,
        description="Statement cache size for asyncpg connections",
    )


settings = Settings()
