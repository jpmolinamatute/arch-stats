#!/usr/bin/env python


import asyncio
import atexit
import logging
import sys
from datetime import datetime
from os import getenv
from uuid import uuid4

import asyncpg
from asyncpg.pool import Pool
from dotenv import load_dotenv

from shared import LogLevel, get_logger

LOGGER = get_logger(__name__, LogLevel.INFO)


class ArchyException(Exception):
    pass


async def get_pg_pool(min_size: int = 1, max_size: int = 10) -> Pool:
    """
    Create and return an asyncpg connection pool.
    Raises an exception if the pool cannot be created.
    """
    try:
        db_pool = await asyncpg.create_pool(
            user=getenv("POSTGRES_USER"),
            database=getenv("POSTGRES_DB"),
            host=getenv("POSTGRES_SOCKET_DIR"),
            password=getenv("POSTGRES_PASSWORD"),
            min_size=min_size,
            max_size=max_size,
        )
        return db_pool
    except Exception as exc:
        raise ArchyException(exc) from exc


async def check_table_exists(db_pool: Pool, table_name: str) -> bool:
    """Check if a table exists in the database."""
    exist = False
    try:
        async with db_pool.acquire() as conn:
            result = await conn.fetchval(
                """
                SELECT to_regclass($1)
                """,
                table_name,
            )

        if result:
            LOGGER.info("✅ Table '%s' exists.", table_name)
            exist = True
        else:
            LOGGER.warning("❌ Table '%s' does not exist.", table_name)
    except Exception as e:
        LOGGER.error("❌ Error checking table '%s': %s", table_name, e)
    return exist


def check_environment_variables() -> None:
    required_env = [
        "POSTGRES_USER",
        "POSTGRES_DB",
        "POSTGRES_PASSWORD",
        "POSTGRES_SOCKET_DIR",
    ]
    missing_env: list[str] = []
    for env in required_env:
        value = getenv(env, None)
        if value is None:
            missing_env.append(env)
    if missing_env:
        msg = f"ERROR: the following environment variables are missing {missing_env}"
        raise ArchyException(msg)


async def check_sessions_table(db_pool: Pool) -> bool:
    """Check if the sessions table exists and if an open session exists, else create one."""
    table_name = "sessions"
    table_exists = await check_table_exists(db_pool, table_name)
    if not table_exists:
        return False

    async with db_pool.acquire() as conn:
        session_count = await conn.fetchval(
            f"""
            SELECT COUNT(id)
            FROM {table_name}
            WHERE is_opened=TRUE;
            """
        )
        if session_count and session_count > 0:
            LOGGER.info("✅ An open session exists.")
            return True

        # No open session, create one

        row = await conn.fetchrow(
            """
            INSERT INTO sessions (is_opened, start_time, location)
            VALUES ($1, $2, $3)
            RETURNING id
            """,
            True,
            datetime.now().astimezone(),
            "default location",
        )
        LOGGER.info("🆕 Created a new open session with id=%s", row["id"])
    return True


async def check_arrows_table(db_pool: Pool) -> bool:
    """Check if the arrows table exists and has data; if not, create 10 arrows."""
    table_name = "arrows"
    table_exists = await check_table_exists(db_pool, table_name)
    if not table_exists:
        return False

    async with db_pool.acquire() as conn:
        arrow_count = await conn.fetchval(
            f"""
            SELECT COUNT(id)
            FROM {table_name};
            """
        )
        if arrow_count and arrow_count > 0:
            LOGGER.info("✅ %d arrows found in the arrows table.", arrow_count)
            return True

        # No arrows found, create 10
        LOGGER.info("🆕 No arrows found. Creating 10 sample arrows.")
        for i in range(10):
            await conn.execute(
                """
                INSERT INTO arrows
                    (id, weight, diameter, spine, length, human_identifier, label_position,
                    is_programmed)
                VALUES
                    ($1, $2, $3, $4, $5, $6, $7, $8)
                """,
                uuid4(),
                20.0,
                6.0,
                400.0,
                28.0,
                f"A{i+1:02d}",
                20.0,
                False,
            )
    return True


async def check_shots_table(db_pool: Pool) -> bool:
    """Check if the shots table exists."""
    return await check_table_exists(db_pool, "shots")


async def pre_start_check(db_pool: Pool) -> bool:
    checks = [
        await check_shots_table(db_pool),
        await check_sessions_table(db_pool),
        await check_arrows_table(db_pool),
    ]
    return all(checks)


async def shooting() -> None:
    pool = await get_pg_pool()
    run_check = True
    while True:
        if run_check:
            if await pre_start_check(pool):
                run_check = False
            else:
                LOGGER.info("Some checks failed. Retrying in 5 seconds...")
                await asyncio.sleep(5)
                continue
        LOGGER.info("Ready")
        await asyncio.sleep(5)


async def main() -> None:
    load_dotenv()
    check_environment_variables()
    await shooting()


if __name__ == "__main__":
    EXIT_CODE = 0
    atexit.register(logging.shutdown)
    try:
        asyncio.run(main())
    except (KeyboardInterrupt, SystemExit):
        LOGGER.info("Bye!")
    except ArchyException as e:
        LOGGER.exception(e)
        EXIT_CODE = 1
    sys.exit(EXIT_CODE)
