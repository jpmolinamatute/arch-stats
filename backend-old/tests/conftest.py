import logging
import time
from collections.abc import AsyncGenerator, Callable, Iterator
from unittest.mock import AsyncMock
from uuid import UUID

import jwt
import pytest
import pytest_asyncio
from app import run
from asyncpg import Pool
from fastapi import FastAPI
from httpx import ASGITransport, AsyncClient

from core import AuthDeps, DBPool, settings
from models import ArcherModel, AuthModel
from routers.v0.auth_router import get_deps


@pytest_asyncio.fixture
async def app() -> AsyncGenerator[FastAPI]:
    application = run()
    # Manually attach expected state for tests (avoid relying on lifespan hooks)
    application.state.logger = logging.getLogger("test")
    application.state.db_pool = await DBPool.open_db_pool()
    try:
        yield application
    finally:
        await DBPool.close_db_pool()


@pytest_asyncio.fixture
async def client(app: FastAPI) -> AsyncGenerator[AsyncClient]:
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test",
        follow_redirects=True,
    ) as ac:
        yield ac


@pytest_asyncio.fixture
async def db_pool(app: FastAPI) -> AsyncGenerator[Pool]:
    # Reuse the pool attached to the app state to avoid lifecycle conflicts
    yield app.state.db_pool


async def _truncate_all(pool: Pool) -> None:
    async with pool.acquire() as conn:
        # Order matters due to FK
        await conn.execute("TRUNCATE shot RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE slot RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE target RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE session RESTART IDENTITY CASCADE")
        await conn.execute("TRUNCATE archer RESTART IDENTITY CASCADE")


@pytest_asyncio.fixture(autouse=True)
async def truncate_db_after_test(db_pool: Pool) -> AsyncGenerator[None]:
    yield
    await _truncate_all(db_pool)


@pytest.fixture
def jwt_for() -> Callable[[UUID], str]:
    def _jwt_for(archer_id: UUID) -> str:
        now = int(time.time())

        return jwt.encode(
            {
                "sub": str(archer_id),
                "iat": now,
                "exp": now + 3600,
                "iss": "arch-stats",
                "typ": "access",
            },
            settings.arch_stats_jwt_secret,
            algorithm=settings.arch_stats_jwt_algorithm,
        )

    return _jwt_for


# Mock Fixtures for individual dependencies
@pytest.fixture
def mock_archers() -> AsyncMock:
    return AsyncMock(spec=ArcherModel)


@pytest.fixture
def mock_sessions() -> AsyncMock:
    return AsyncMock(spec=AuthModel)


# Override the AuthDeps dependency using the mock fixtures
@pytest.fixture
def setup_auth_deps(
    app: FastAPI, mock_archers: AsyncMock, mock_sessions: AsyncMock
) -> Iterator[None]:
    deps = AuthDeps(archers=mock_archers, sessions=mock_sessions)
    app.dependency_overrides[get_deps] = lambda: deps
    yield
    app.dependency_overrides = {}
