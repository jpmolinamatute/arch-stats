from collections.abc import AsyncGenerator

import pytest_asyncio
from asyncpg import Pool
from httpx import ASGITransport, AsyncClient

from server import create_app, create_tables
from server.models import ArrowsDB, DBState, SessionsDB, ShotsDB, TargetsDB
from shared import LogLevel, get_logger


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
    await create_tables()
    logger = get_logger("test", LogLevel.DEBUG)
    app = create_app(logger)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        yield ac
    await drop_tables()
    await DBState.close_db()


@pytest_asyncio.fixture
async def db_pool() -> AsyncGenerator[Pool, None]:
    pool = await DBState.get_db_pool()
    yield pool
