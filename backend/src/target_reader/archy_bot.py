#!/usr/bin/env python
import asyncio
import logging
import random
import sys
from datetime import datetime, timedelta
from uuid import UUID, uuid4

from asyncpg import Pool

from shared.db_pool import DBPool
from shared.logger import LogLevel, get_logger
from shared.models import ArrowsDB, SessionsDB, ShotsDB
from shared.schema import ArrowsCreate, SessionsCreate, ShotsCreate


class ArchyException(Exception):
    pass


class ArchyBot:
    def __init__(self, logger: logging.Logger, pool: Pool) -> None:
        self.logger = logger
        self.pool = pool
        self.arrows_db = ArrowsDB(self.pool)
        self.shots_db = ShotsDB(self.pool)
        self.sessions_db = SessionsDB(self.pool)

    async def start(self) -> None:
        try:
            await self.shooting()
        except Exception as e:
            self.logger.exception("Error in shooting: %s", e)
        finally:
            await self.close()

    async def insert_new_session(self) -> UUID:
        payload = SessionsCreate(
            is_opened=True,
            start_time=datetime.now().astimezone(),
            location="default location",
            is_indoor=False,
            distance=18,
        )
        _id = await self.sessions_db.insert_one(payload)
        self.logger.info("A session was inserted in the DB")
        return _id

    async def get_open_session_id(self) -> UUID:
        session = await self.sessions_db.get_open_session()
        if session is None:
            session_id = await self.insert_new_session()
        else:
            session_id = session.get_id()
        return session_id

    async def insert_new_arrows(self) -> list[UUID]:
        """Insert 10 arrows using ArrowsDB.insert_one and return their UUIDs."""

        arrow_uuids: list[UUID] = []
        now = datetime.now().astimezone()
        for i in range(10):
            new_id = uuid4()
            payload = ArrowsCreate(
                id=new_id,
                human_identifier=f"A{i+1:02d}",
                length=28.0,
                registration_date=now,
                is_programmed=False,
                is_active=True,
                voided_date=None,
                weight=20.0,
                diameter=6.0,
                spine=400.0,
                label_position=20.0,
            )
            _id = await self.arrows_db.insert_one(payload)
            arrow_uuids.append(_id)

        self.logger.info("10 arrows were inserted in the DB")
        return arrow_uuids

    async def get_all_arrow_ids(self) -> list[UUID]:
        rows = await self.arrows_db.get_all({"is_active": True})
        if rows:
            return [row.get_id() for row in rows]
        return await self.insert_new_arrows()

    async def insert_shot(self, session_id: UUID, arrows_ids: list[UUID]) -> None:
        """Insert a shot using the shared ShotsDB model.

        Chooses a random arrow id, generates times and optional landing/x/y,
        then persists via ShotsDB.insert_one.
        """
        arrow_id = random.choice(arrows_ids)
        now = datetime.now().astimezone()
        disengage_time = now + timedelta(seconds=4)
        if random.choice([True, False]):
            landing_time = None
            x = None
            y = None
        else:
            landing_time = now + timedelta(seconds=6)
            x = random.uniform(0.0, 100.0)
            y = random.uniform(0.0, 100.0)

        payload = ShotsCreate(
            arrow_id=arrow_id,
            session_id=session_id,
            arrow_engage_time=now,
            arrow_disengage_time=disengage_time,
            arrow_landing_time=landing_time,
            x=x,
            y=y,
        )
        await self.shots_db.insert_one(payload)
        self.logger.info("A shot was inserted in the DB")

    async def pre_start_check(self) -> bool:
        """Ensure required DB objects exist; create them if missing.

        When the bot runs standalone (without the FastAPI server creating tables
        on startup), the necessary tables may not exist yet. This routine
        verifies existence and creates what's missing so the bot can run on its
        own during development.
        """
        try:
            # Ensure required extension exists (uuid-ossp for uuid_generate_v4)
            async with self.pool.acquire() as conn:
                await conn.execute('CREATE EXTENSION IF NOT EXISTS "uuid-ossp";')

            # Check each table; create if needed
            if not await self.sessions_db.check_table_exists():
                await self.sessions_db.create_table()
                self.logger.info("Created 'sessions' table")
            if not await self.arrows_db.check_table_exists():
                await self.arrows_db.create_table()
                self.logger.info("Created 'arrows' table")
            if not await self.shots_db.check_table_exists():
                await self.shots_db.create_table()
                self.logger.info("Created 'shots' table")

            # Ensure notifications are present for shots (uses default channel from settings via
            # server in normal run; here adopt a simple channel name)
            try:
                # Avoid hard dependency on settings here; use a default dev channel
                await self.shots_db.create_notification("archy")
            except Exception as e:
                # Non-fatal for the bot operation
                self.logger.debug("Skipping shots notification setup: %s", e)

            return True
        except Exception as e:
            self.logger.warning("Pre-start check failed: %s", e)
            return False

    async def shooting(self) -> None:
        run_check = True
        arrows_ids: list[UUID] = []
        session_id: UUID | None = None
        backoff_delay: float = 1.0
        max_backoff: float = 30.0

        while True:
            if run_check:
                if await self.pre_start_check():
                    run_check = False
                    # we initialize session_id and arrows_id only once and only
                    # after pre_start_check() is True
                    session_id = await self.get_open_session_id()
                    arrows_ids = await self.get_all_arrow_ids()
                else:
                    self.logger.info("Some checks failed. Retrying in 5 seconds...")
                    await asyncio.sleep(5)
                    continue

            # Explicit runtime guard (asserts can be stripped with -O)
            if session_id is None or not arrows_ids:
                raise ArchyException("Session/Arrows IDs not initialized!")

            # Resilience: attempt to insert a shot, retrying with exponential
            # backoff and jitter on transient failures instead of crashing.
            try:
                await self.insert_shot(session_id, arrows_ids)
                # Success: reset backoff
                backoff_delay = 1.0
            except Exception as e:
                # Log and backoff with jitter; keep the bot alive
                self.logger.warning(
                    "Shot insertion failed: %s. Backing off for ~%.1fs before retrying.",
                    e,
                    backoff_delay,
                )
                jitter = random.uniform(0, backoff_delay / 2)
                await asyncio.sleep(backoff_delay + jitter)
                backoff_delay = min(max_backoff, backoff_delay * 2)
                # Skip normal reload sleep this iteration and retry
                continue

            time_between_shots = random.randint(10, 15)
            self.logger.info("Reloading bow")
            await asyncio.sleep(time_between_shots)

    async def close(self) -> None:
        await DBPool.close_db_pool()
        self.logger.info("Database pool closed.")


async def run() -> None:
    logger = get_logger(file_name=__name__, log_level=LogLevel.INFO)
    pool = await DBPool.open_db_pool()
    app = ArchyBot(logger, pool)
    try:
        await app.start()
    except asyncio.exceptions.CancelledError:
        pass
    except ArchyException as e:
        logger.exception(e)
    except Exception as e:
        logger.exception("Unexpected error: %s", e)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    EXIT_STATUS = 0
    try:
        asyncio.run(run())
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(e)
        EXIT_STATUS = 2
    finally:
        sys.exit(EXIT_STATUS)
