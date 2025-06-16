from asyncpg import Pool, create_pool

from server.settings import settings


class DBStateError(Exception):
    pass


class DBState:
    db_pool: Pool | None = None

    @classmethod
    async def init_db(cls) -> None:
        """Create and store the database connection pool."""
        if cls.db_pool is None:
            params = {
                "user": settings.postgres_user,
                "database": settings.postgres_db,
                "host": settings.postgres_socket_dir,
                "min_size": 1,
                "max_size": 10,
                "max_queries": 50000,
                "max_inactive_connection_lifetime": 300,
            }
            cls.db_pool = await create_pool(**params)

    @classmethod
    async def close_db(cls) -> None:
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
