from os import getenv

from asyncpg import Pool, create_pool


class DBStateError(Exception):
    pass


class DBState:
    db_pool: Pool | None = None

    @classmethod
    def check_envs(cls) -> None:
        env_to_check = [
            "POSTGRES_USER",
            "POSTGRES_DB",
            "POSTGRES_SOCKET_DIR",
        ]
        missing_env = []
        for env in env_to_check:
            if getenv(env, None) is None:
                missing_env.append(env)
        if missing_env:
            message = ", ".join(missing_env)
            raise DBStateError(f"ERROR: environment variables missing '{message}'")

    @classmethod
    async def init_db(cls) -> None:
        """Create and store the database connection pool."""
        cls.check_envs()
        if cls.db_pool is None:
            params = {
                "user": getenv("POSTGRES_USER"),
                "database": getenv("POSTGRES_DB"),
                "host": getenv("POSTGRES_SOCKET_DIR"),
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
