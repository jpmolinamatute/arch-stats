import asyncio
from pathlib import Path
from typing import Any, Self

from asyncpg import Pool, create_pool

from core.settings import settings


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
        """
        Get the socket or hostname for the database connection.

        Preference: use Unix socket if the actual socket file exists; otherwise
        use TCP host if provided. This avoids selecting a socket based solely on
        the directory existing (common on CI images) when no server is listening.
        """
        # Prefer Unix socket only when the socket file exists for the configured port
        if settings.postgres_socket_dir:
            socket_dir = Path(settings.postgres_socket_dir)
            socket_file = socket_dir / f".s.PGSQL.{settings.postgres_port}"
            if socket_dir.is_dir() and socket_file.exists():
                return settings.postgres_socket_dir

        # Fallback to TCP host if set
        if settings.postgres_host:
            return settings.postgres_host

        raise DBStateError(
            "Database connection requires either an active Unix socket at "
            f"{settings.postgres_socket_dir!r} or a TCP host via 'postgres_host'."
        )

    @classmethod
    def _get_lock(cls) -> asyncio.Lock:
        """Lazily create and return the class-level lock (requires running loop)."""
        if cls._lock is None:
            cls._lock = asyncio.Lock()
        return cls._lock

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
            max_inactive_connection_lifetime = settings.postgres_max_inactive_connection_lifetime
            # We are ignoring a false positive, cls._get_lock() always returns a asyncio.Lock
            async with cls._get_lock():  # pylint: disable=not-async-context-manager
                cls._pool = await create_pool(
                    user=settings.postgres_user,
                    database=settings.postgres_db,
                    password=settings.postgres_password,
                    host=cls._socket_or_host(),
                    port=settings.postgres_port,
                    min_size=settings.postgres_pool_min_size,
                    max_size=settings.postgres_pool_max_size,
                    max_queries=settings.postgres_max_queries,
                    max_inactive_connection_lifetime=max_inactive_connection_lifetime,
                    command_timeout=settings.postgres_command_timeout,
                    statement_cache_size=settings.postgres_statement_cache_size,
                )
        assert isinstance(cls._pool, Pool)
        return cls._pool
