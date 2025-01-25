# dependencies.py
from asyncpg import Pool

from server.db import DBState


async def get_db() -> Pool:
    """Dependency to get the PostgreSQL connection pool."""
    if DBState.db_pool is None:
        raise RuntimeError("Database connection pool is not initialized")
    return DBState.db_pool
