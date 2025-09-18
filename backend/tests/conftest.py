import logging
from collections.abc import AsyncGenerator

import pytest_asyncio
from asyncpg import Pool
from fastapi import FastAPI
from httpx import ASGITransport, AsyncClient

from app import run
from core import DBPool


# pylint: disable=redefined-outer-name


@pytest_asyncio.fixture
async def app() -> AsyncGenerator[FastAPI, None]:
    application = run()
    # Manually attach expected state for tests (avoid relying on lifespan hooks)
    application.state.logger = logging.getLogger("test")
    application.state.db_pool = await DBPool.open_db_pool()
    try:
        yield application
    finally:
        await DBPool.close_db_pool()


@pytest_asyncio.fixture
async def client(app: FastAPI) -> AsyncGenerator[AsyncClient, None]:
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test",
        follow_redirects=True,
    ) as ac:
        yield ac


@pytest_asyncio.fixture
async def db_pool(app: FastAPI) -> AsyncGenerator[Pool, None]:
    # Reuse the pool attached to the app state to avoid lifecycle conflicts
    yield app.state.db_pool
