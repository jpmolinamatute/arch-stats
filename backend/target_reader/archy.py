#!/usr/bin/env python
#!/usr/bin/env python
import asyncio
import logging
import random
import sys
from datetime import datetime, timedelta
from os import getenv
from uuid import UUID, uuid4

import asyncpg
from asyncpg.connection import Connection
from asyncpg.pool import Pool

from shared import LogLevel, get_logger


class ArchyException(Exception):
    pass


class ArchyApp:
    def __init__(self, logger: logging.Logger) -> None:
        self.logger = logger
        self.db_pool: Pool | None = None

    async def start(self) -> None:
        self.check_environment_variables()
        self.db_pool = await self.get_pg_pool()
        try:
            await self.shooting()
        finally:
            await self.close_db_pool()

    async def get_pg_pool(self, min_size: int = 1, max_size: int = 10) -> Pool:
        try:
            db_pool = await asyncpg.create_pool(
                user=getenv("POSTGRES_USER"),
                database=getenv("POSTGRES_DB"),
                host=getenv("POSTGRES_SOCKET_DIR"),
                password=getenv("POSTGRES_PASSWORD"),
                min_size=min_size,
                max_size=max_size,
            )
            self.logger.info("Database pool created")
            return db_pool
        except Exception as exc:
            raise ArchyException(exc) from exc

    def check_environment_variables(self) -> None:
        required_env = [
            "POSTGRES_USER",
            "POSTGRES_DB",
            "POSTGRES_PASSWORD",
            "POSTGRES_SOCKET_DIR",
        ]
        missing_env = []
        for env in required_env:
            value = getenv(env, None)
            if value is None:
                missing_env.append(env)
        if missing_env:
            msg = f"ERROR: the following environment variables are missing {missing_env}"
            raise ArchyException(msg)

    async def check_table_exists(self, conn: Connection, table_name: str) -> bool:
        exist = False
        try:
            result = await conn.fetchval("SELECT to_regclass($1)", table_name)
            if result:
                self.logger.info("Table '%s' exists.", table_name)
                exist = True
            else:
                self.logger.warning("Table '%s' does not exist.", table_name)
        except Exception as e:
            self.logger.error("Error checking table '%s': %s", table_name, e)
        return exist

    async def insert_new_session(self, conn: Connection) -> UUID:
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
        self.logger.info("A session was inserted in the DB")
        if row is None:
            raise ArchyException("ERROR: we couldn't insert a session")
        _id: UUID = row["id"]
        return _id

    async def get_open_session_id(self, conn: Connection) -> UUID:
        session_id: UUID | None = await conn.fetchval(
            """
            SELECT id
            FROM sessions
            WHERE is_opened=TRUE
            LIMIT 1;
            """
        )
        if session_id is None:
            session_id = await self.insert_new_session(conn)
        return session_id

    async def insert_new_arrows(self, conn: Connection) -> list[UUID]:
        """
        Insert 10 arrows into the arrows table and return their UUIDs.
        Uses executemany for efficient batch insertion.
        """
        arrow_uuids = [uuid4() for _ in range(10)]
        records = [
            (
                arrow_id,
                20.0,
                6.0,
                400.0,
                28.0,
                f"A{i+1:02d}",
                20.0,
                False,
            )
            for i, arrow_id in enumerate(arrow_uuids)
        ]
        query = """
            INSERT INTO arrows (
                id, weight, diameter, spine, length,
                human_identifier, label_position, is_programmed
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        """
        await conn.executemany(query, records)
        self.logger.info("10 arrows were inserted in the DB")
        return arrow_uuids

    async def get_all_arrow_ids(self, conn: Connection) -> list[UUID]:
        rows = await conn.fetch(
            """
            SELECT id
            FROM arrows;
            """
        )
        uuid_list = []
        if rows:
            uuid_list = [row["id"] for row in rows]
        else:
            uuid_list = await self.insert_new_arrows(conn)
        return uuid_list

    async def check_sessions_table(self, conn: Connection) -> bool:
        return await self.check_table_exists(conn, "sessions")

    async def check_arrows_table(self, conn: Connection) -> bool:
        return await self.check_table_exists(conn, "arrows")

    async def check_shots_table(self, conn: Connection) -> bool:
        return await self.check_table_exists(conn, "shots")

    async def insert_shot(self, conn: Connection, session_id: UUID, arrows_ids: list[UUID]) -> None:
        arrow_id = random.choice(arrows_ids)
        now = datetime.now().astimezone()
        disengage_time = now + timedelta(seconds=4)
        # Randomly decide whether landing_time is None or now + 6 seconds
        if random.choice([True, False]):
            landing_time = None
            x_coordinate = None
            y_coordinate = None
        else:
            landing_time = now + timedelta(seconds=6)
            x_coordinate = random.uniform(0.0, 100.0)
            y_coordinate = random.uniform(0.0, 100.0)
        await conn.execute(
            """
            INSERT INTO shots (
                arrow_id,
                session_id,
                arrow_engage_time,
                arrow_disengage_time,
                arrow_landing_time,
                x_coordinate,
                y_coordinate
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7);
            """,
            arrow_id,
            session_id,
            now,
            disengage_time,
            landing_time,
            x_coordinate,
            y_coordinate,
        )
        self.logger.info("A shot was inserted in the DB")

    async def pre_start_check(self, conn: Connection) -> bool:
        checks = [
            await self.check_shots_table(conn),
            await self.check_sessions_table(conn),
            await self.check_arrows_table(conn),
        ]
        return all(checks)

    async def shooting(self) -> None:
        run_check = True
        arrows_ids: list[UUID] = []
        session_id: UUID | None = None
        assert self.db_pool is not None
        async with self.db_pool.acquire() as conn:
            while True:
                if run_check:
                    if await self.pre_start_check(conn):
                        run_check = False
                        # we initialized session_id and arrows_id only once and only
                        # after self.pre_start_check(conn) is True
                        session_id = await self.get_open_session_id(conn)
                        arrows_ids = await self.get_all_arrow_ids(conn)
                    else:
                        self.logger.info("Some checks failed. Retrying in 5 seconds...")
                        await asyncio.sleep(5)
                        continue

                assert session_id is not None and arrows_ids, "Session/Arrows IDs not initialized!"
                await self.insert_shot(conn, session_id, arrows_ids)
                time_between_shoots = random.randint(10, 15)
                self.logger.info("Reloading bow")
                await asyncio.sleep(time_between_shoots)

    async def close_db_pool(self) -> None:
        if self.db_pool:
            await self.db_pool.close()
            self.logger.info("Database pool closed.")


async def run() -> None:
    logger = get_logger(file_name=__name__, log_lever=LogLevel.INFO)
    app = ArchyApp(logger)
    try:
        await app.start()
    except asyncio.exceptions.CancelledError:
        pass
    except ArchyException as e:
        logger.exception(e)
    except Exception as e:
        logger.exception("Unexpected error: %s", e)
    finally:
        await app.close_db_pool()


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
