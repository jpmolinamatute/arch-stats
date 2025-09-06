import logging
from collections.abc import AsyncGenerator

import pytest_asyncio
from asyncpg import Pool
from httpx import ASGITransport, AsyncClient

from server.app import lifespan, manage_tables, run
from shared.db_pool import DBPool


@pytest_asyncio.fixture
async def async_client() -> AsyncGenerator[AsyncClient, None]:
    """Provide an AsyncClient bound to a fresh app lifecycle per test."""
    logging.basicConfig(level=logging.DEBUG)
    app = run()
    app.state.logger = logging.getLogger("tests")
    async with lifespan(app):
        await manage_tables(app.state.db_pool, "drop")
        await manage_tables(app.state.db_pool, "create")
        transport = ASGITransport(app=app)
        async with AsyncClient(
            transport=transport,
            base_url="http://test",
        ) as ac:
            yield ac


@pytest_asyncio.fixture
async def db_pool() -> AsyncGenerator[Pool, None]:
    yield await DBPool.open_db_pool()
    await DBPool.close_db_pool()


@pytest_asyncio.fixture
async def db_pool_initialed() -> AsyncGenerator[Pool, None]:
    """DB pool with schema created; cleans up tables and closes the pool after use."""
    pool = await DBPool.open_db_pool()
    await manage_tables(pool, "create")
    yield pool
    await manage_tables(pool, "drop")
    await DBPool.close_db_pool()
