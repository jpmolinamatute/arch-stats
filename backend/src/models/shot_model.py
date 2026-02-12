import asyncio
import json
from collections.abc import AsyncIterator
from uuid import UUID

from asyncpg import Connection, Pool
from asyncpg.pool import PoolConnectionProxy

from models.parent_model import DBNotFound, ParentModel
from schema import LiveStat, ShotCreate, ShotFilter, ShotRead, ShotScore, ShotSet, Stats


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
        where = ShotFilter(slot_id=slot_id)
        query, params = self.build_select_view_sql_stm(
            view_name="live_stat_by_slot_id",
            where=where,
            columns=["*"],
            limit=1,
            is_desc=False,
        )
        result: LiveStat
        try:
            row = await self.fetchrow((query, params))
        except DBNotFound:
            row = None

        if row is None:
            result = LiveStat(
                slot_id=slot_id,
                number_of_shots=0,
                total_score=0,
                max_score=0,
                mean=0.0,
            )
        else:
            result = LiveStat(
                slot_id=row["slot_id"],
                number_of_shots=row["number_of_shots"],
                total_score=row["total_score"],
                max_score=row["max_score"],
                mean=row["mean"],
            )

        return result

    async def get_scores(self, slot_id: UUID) -> list[ShotScore]:
        """Retrieve all scores for a given slot."""
        where = ShotFilter(slot_id=slot_id)
        query, params = self.build_select_sql_stm(
            where=where,
            columns=["shot_id", "score", "is_x", "created_at"],
            limit=0,
            is_desc=False,
        )
        rows = await self.fetch((query, params))
        return [ShotScore(**dict(row)) for row in rows]

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
        queue: asyncio.Queue[list[ShotScore]] = asyncio.Queue()

        def _listener(
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
            try:
                parsed = json.loads(payload_str)
            except json.JSONDecodeError:
                self.logger.warning("Invalid JSON payload received: %s", payload_str)
                return

            raw_items = []
            if isinstance(parsed, list):
                raw_items = parsed
            elif isinstance(parsed, dict):
                raw_items = [parsed]

            shot_scores: list[ShotScore] = []
            for item in raw_items:
                try:
                    # We first validate against ShotRead to ensure we have a valid shot record
                    # payload usually comes from the DB table row.
                    shot_read = ShotRead(**item)
                    if shot_read.score is not None:
                        shot_scores.append(
                            ShotScore(
                                shot_id=shot_read.shot_id,
                                score=shot_read.score,
                                is_x=shot_read.is_x,
                                created_at=shot_read.created_at,
                            )
                        )
                except Exception:
                    # Skip invalid items
                    continue

            if shot_scores:
                queue.put_nowait(shot_scores)

        async with self.db_pool.acquire() as conn:
            await conn.add_listener(channel_name, _listener)
            try:
                while True:
                    shots_score = await queue.get()
                    # Now we are in the main loop, we can safely await async calls
                    current_stats = await self.get_live_stat(slot_id)

                    stats = Stats(
                        shots=shots_score,
                        stats=current_stats,
                    )
                    yield stats
            finally:
                await conn.remove_listener(channel_name, _listener)
