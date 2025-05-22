import os
from collections.abc import AsyncGenerator, Generator

import pytest
import pytest_asyncio
from asyncpg import Connection, Pool, connect, create_pool
from fastapi.testclient import TestClient

from server.models import ArrowsDB, SessionsDB, ShotsDB, TargetsDB
from server import create_app

# pylint: disable=redefined-outer-name


@pytest.fixture
def client() -> Generator[TestClient, None, None]:
    app = create_app()
    with TestClient(app) as c:
        yield c


@pytest_asyncio.fixture(loop_scope="session")
async def test_db() -> AsyncGenerator[Pool, None]:
    """
    Creates a fresh test database, runs your DB classes' create_table methods,
    and yields a connection pool.
    Drops the test database at teardown.
    """

    postgres_user = os.getenv("POSTGRES_USER")
    # postgres_password = os.getenv("POSTGRES_PASSWORD")
    postgres_db = os.getenv("POSTGRES_DB")
    postgres_host = os.getenv("POSTGRES_SOCKET_DIR")
    test_db_name = "testdb"

    admin_dsn = {
        "host": postgres_host,
        "user": postgres_user,
        "database": postgres_db,
        "timeout": 60,
        "statement_cache_size": 100,
        "max_cached_statement_lifetime": 300,
        "max_cacheable_statement_size": 1024 * 15,
    }
    testdb_dsn = {
        "host": postgres_host,
        "user": postgres_user,
        "database": test_db_name,
        "timeout": 60,
        "statement_cache_size": 100,
        "max_cached_statement_lifetime": 300,
        "max_cacheable_statement_size": 1024 * 15,
    }

    # 1. Connect to postgres (default db), create test db if not exists
    admin_conn = await connect(**admin_dsn)
    try:
        print("Dropping previous DB")
        await admin_conn.execute(f'DROP DATABASE IF EXISTS "{test_db_name}";')
        print(f"Creating new DB '{test_db_name}'")
        await admin_conn.execute(f'CREATE DATABASE "{test_db_name}";')
    finally:
        await admin_conn.close()

    # 2. Connect to the new test database and create schema via your DB classes
    pool = await create_pool(**testdb_dsn)
    # Create tables using your real classes
    async with pool.acquire():
        # Use the pool, not conn, to init your DB classes (so each has access to the pool)
        print("Creating Tables")
        await ArrowsDB(pool).create_table()
        await SessionsDB(pool).create_table()
        await ShotsDB(pool).create_table()
        await TargetsDB(pool).create_table()
    yield pool

    # 3. Drop the test db at teardown (optional but usually good)
    await pool.close()

    admin_conn = await connect(admin_dsn)
    try:
        print(f"Drooping DB '{test_db_name}'")
        await admin_conn.execute(f'DROP DATABASE IF EXISTS "{test_db_name}";')
    finally:
        await admin_conn.close()
        print("Teardown complete")


@pytest_asyncio.fixture
async def db_conn(test_db: Pool) -> AsyncGenerator[Connection, None]:
    """Get a connection from the pool for each test."""
    async with test_db.acquire() as conn:
        yield conn
