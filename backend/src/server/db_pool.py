from pathlib import Path
from typing import Any, Self

from asyncpg import Pool, create_pool

from server.settings import settings


class DBStateError(Exception):
    pass


class DBPool:
    """Borg class to manage the PostgreSQL database connection pool."""

    db_pool: Pool | None = None

    def __new__(cls, *args: Any, **kwargs: Any) -> Self:
        raise TypeError("DBPool should not be instantiated. Use class methods only.")

    @staticmethod
    def socket_or_host() -> str:
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

    @staticmethod
    def create_pool_params() -> dict[str, Any]:
        """Build parameters for asyncpg.create_pool().

        - If connecting via UNIX socket (host is a directory), do not include password.
        - If connecting via hostname/IP, include password from settings.postgres_password.
        """
        host_value = DBPool.socket_or_host()
        is_unix_socket = Path(host_value).is_dir()

        params: dict[str, Any] = {
            "user": settings.postgres_user,
            "database": settings.postgres_db,
            "host": host_value,
            "port": settings.postgres_port,
            "min_size": 1,
            "max_size": 10,
            "max_queries": 50000,
            "max_inactive_connection_lifetime": 300,
        }

        if not is_unix_socket:
            params["password"] = settings.postgres_password

        return params

    @classmethod
    async def create_db_pool(cls) -> None:
        """Create and store the database connection pool."""
        if cls.db_pool is None:
            params = cls.create_pool_params()
            cls.db_pool = await create_pool(**params)

    @classmethod
    async def close_db_pool(cls) -> None:
        """Close the database connection pool on shutdown."""
        if cls.db_pool is not None:
            await cls.db_pool.close()
            cls.db_pool = None

    @classmethod
    async def get_db_pool(cls) -> Pool:
        """Dependency to get the PostgreSQL connection pool."""
        if cls.db_pool is None:
            raise RuntimeError("Database connection pool is not initialized")
        return cls.db_pool
