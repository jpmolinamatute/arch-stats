from collections.abc import AsyncGenerator

import pytest_asyncio
from asyncpg import Pool
from httpx import ASGITransport, AsyncClient

from server.app import create_tables, run
from server.models import ArrowsDB, DBState, SessionsDB, ShotsDB, TargetsDB


async def drop_tables() -> None:
    pool = await DBState.get_db_pool()
    arrows = ArrowsDB(pool)
    shots = ShotsDB(pool)
    sessions = SessionsDB(pool)
    targets = TargetsDB(pool)
    await targets.drop_table()
    await shots.drop_table()
    await arrows.drop_table()
    await sessions.drop_table()


@pytest_asyncio.fixture
async def async_client() -> AsyncGenerator[AsyncClient, None]:
    await DBState.init_db()
    pool = await DBState.get_db_pool()
    await create_tables(pool)
    app = run()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        yield ac
    await drop_tables()
    await DBState.close_db()


@pytest_asyncio.fixture
async def db_pool() -> AsyncGenerator[Pool, None]:
    pool = await DBState.get_db_pool()
    yield pool


@pytest_asyncio.fixture
async def db_pool_initialed() -> AsyncGenerator[Pool, None]:
    await DBState.init_db()
    pool = await DBState.get_db_pool()
    await create_tables(pool)
    yield pool
    await drop_tables()
    await DBState.close_db()
