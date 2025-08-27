import asyncio
from pathlib import Path
from typing import Any, Self

from asyncpg import Pool, create_pool

from shared.settings import settings


class DBStateError(Exception):
    """DB-related state error."""


class DBPool:
    """Borg class to manage the PostgreSQL database connection pool."""

    _pool: Pool | None = None
    _lock: asyncio.Lock | None = None

    def __new__(cls, *args: Any, **kwargs: Any) -> Self:
        raise TypeError("DBPool should not be instantiated. Use class methods only.")

    @classmethod
    def _socket_or_host(cls) -> str:
        """Get the socket or host for the database connection."""
        base_value = ""
        if settings.postgres_socket_dir and Path(settings.postgres_socket_dir).is_dir():
            base_value = settings.postgres_socket_dir
        elif settings.postgres_host:
            base_value = settings.postgres_host
        else:
            raise DBStateError(
                "Database connection requires either a valid socket directory (postgres_socket_dir)"
                " or host (postgres_host) to be configured."
            )
        return base_value

    @classmethod
    def _get_lock(cls) -> asyncio.Lock:
        """Lazily create and return the class-level lock (requires running loop)."""
        if cls._lock is None:
            cls._lock = asyncio.Lock()
        return cls._lock

    @classmethod
    def _create_pool_params(cls) -> dict[str, Any]:
        """Build parameters for asyncpg.create_pool()."""
        host_value = cls._socket_or_host()
        is_unix_socket = Path(host_value).is_dir()

        params: dict[str, Any] = {
            "user": settings.postgres_user,
            "database": settings.postgres_db,
            "host": host_value,
            "port": settings.postgres_port,
            "min_size": settings.pg_pool_min,
            "max_size": settings.pg_pool_max,
            "max_queries": 50_000,
            "max_inactive_connection_lifetime": 300,
            "command_timeout": settings.command_timeout,
            "statement_cache_size": settings.statement_cache_size,
        }

        if not is_unix_socket:
            params["password"] = settings.postgres_password

        return params

    @classmethod
    async def close_db_pool(cls) -> None:
        """Close the database connection pool on shutdown (race-safe)."""

        if cls._pool is not None:
            # We are ignoring a false positive, cls._get_lock() always returns a asyncio.Lock
            async with cls._get_lock():  # pylint: disable=not-async-context-manager
                await cls._pool.close()
                cls._pool = None

    @classmethod
    async def open_db_pool(cls) -> Pool:
        """Return store the database connection pool (race-safe)."""

        if cls._pool is None:
            # We are ignoring a false positive, cls._get_lock() always returns a asyncio.Lock
            async with cls._get_lock():  # pylint: disable=not-async-context-manager
                cls._pool = await create_pool(**cls._create_pool_params())
        return cls._pool
