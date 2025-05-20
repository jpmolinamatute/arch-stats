from os import getenv

from asyncpg import Pool, create_pool


class DBState:
    db_pool: Pool | None = None

    @classmethod
    async def init_db(cls) -> None:
        """Create and store the database connection pool."""
        if cls.db_pool is None:
            user = getenv("ARCH_STATS_USER")
            params = {}
            if user:
                params["user"] = user
            cls.db_pool = await create_pool(**params, min_size=1, max_size=10)

    @classmethod
    async def close_db(cls) -> None:
        """Close the database connection pool on shutdown."""
        if cls.db_pool is not None:
            await cls.db_pool.close()
            cls.db_pool = None
