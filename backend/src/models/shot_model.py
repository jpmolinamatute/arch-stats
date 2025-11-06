import asyncio
import json
from collections.abc import AsyncIterator
from typing import TypedDict
from uuid import UUID

from asyncpg import Connection, Pool
from asyncpg.pool import PoolConnectionProxy

from models.parent_model import ParentModel
from schema import ShotCreate, ShotFilter, ShotRead, ShotSet


class ShotScore(TypedDict):
    shot_id: UUID
    score: int


class LiveStat(TypedDict):
    slot_id: UUID
    number_of_shots: int
    total_score: int
    max_score: int
    mean: float


class Stats(TypedDict):
    shot: ShotScore
    stats: LiveStat


class ShotModel(ParentModel[ShotCreate, ShotSet, ShotRead, ShotFilter]):
    """Model for shot-related DB access and notifications."""

    def __init__(self, db_pool: Pool) -> None:
        super().__init__("shot", db_pool, ShotRead)

    async def get_live_stat(self, slot_id: UUID) -> LiveStat:
        """Retrieve live statistics for shots in a given slot.

        Uses the materialized view `live_stat_by_slot_id` which aggregates
        shots per slot. If no row exists for the provided slot (no shots yet),
        returns zeros.
        """
        query = """
            SELECT *
            FROM live_stat_by_slot_id
            WHERE slot_id = $1
            LIMIT 1;
            """
        result: LiveStat
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching live stats: %s", query)
            row = await conn.fetchrow(query, slot_id)

        if row is None:
            result = {
                "slot_id": slot_id,
                "number_of_shots": 0,
                "total_score": 0,
                "max_score": 0,
                "mean": 0.0,
            }
        else:
            result = {
                "slot_id": row["slot_id"],
                "number_of_shots": row["number_of_shots"],
                "total_score": row["total_score"],
                "max_score": row["max_score"],
                "mean": row["mean"],
            }

        return result

    async def listen_for_shots(self, slot_id: UUID) -> AsyncIterator[Stats]:
        """Yield shot notifications for a specific slot.

        Contract:
        - Input: slot_id (UUID) identifies the slot to listen on.
        - Output: async iterator yielding parsed payloads (JSON-decoded when possible,
          otherwise raw strings) coming from the LISTEN/NOTIFY channel
          f"{self.name}_insert_{slot_id}".
        - Cleanup: listener is removed when the consumer stops iterating or on
          cancellation; no global state is left behind.
        """

        channel_name = f"{self.name}_insert_{slot_id}"
        queue: asyncio.Queue[Stats] = asyncio.Queue()

        async def _listener(
            _: Connection | PoolConnectionProxy,
            __: int,
            ___: str,
            payload: object,
        ) -> None:
            """
            Push incoming payloads into an async queue for consumption.

            The `payload` from asyncpg notifications is a string. We attempt to
            JSON-decode it; if it looks like a dict containing `shot_id` and
            `score`.
            """

            payload_str = payload if isinstance(payload, str) else str(payload)

            parsed = json.loads(payload_str)
            # Check for required keys if payload is a JSON object
            missing_keys = [k for k in ("shot_id", "score") if k not in parsed]
            if missing_keys:
                raise ValueError(f"Invalid payload structure: missing keys {missing_keys}")

            stats: Stats = {
                "shot": {
                    "shot_id": UUID(parsed["shot_id"]),
                    "score": parsed["score"],
                },
                "stats": await self.get_live_stat(slot_id),
            }
            await queue.put(stats)

        async with self.db_pool.acquire() as conn:
            await conn.add_listener(channel_name, _listener)
            try:
                while True:
                    item: Stats = await queue.get()
                    yield item
            finally:
                await conn.remove_listener(channel_name, _listener)
