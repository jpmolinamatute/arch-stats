import logging
from collections.abc import AsyncGenerator
from uuid import UUID

from fastapi import HTTPException, status

from core.base_manager import BaseManager
from models.live_stats_model import LiveStatsModel
from models.parent_model import DBNotFound
from schema.live_stats_schema import LiveStat


class LiveStatsManager(BaseManager):
    def __init__(self, db_pool, logger: logging.Logger):
        super().__init__(db_pool)
        self.live_stats_model = LiveStatsModel(db_pool)
        self.logger = logger

    async def get_stats(self, slot_id: UUID, current_archer_id: UUID) -> LiveStat:
        try:
            await self.verify_slot_ownership(slot_id, current_archer_id)
            live_stat = await self.live_stats_model.get_live_stat(slot_id)
            scores = await self.live_stats_model.get_all_scores(slot_id)
            return LiveStat(scores=scores, stats=live_stat)
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
        except HTTPException:
            raise
        except Exception as e:
            self.logger.error("Internal Server Error in get_stats: %s", str(e), exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail="Internal Server Error"
            ) from e

    async def listen_for_shots(self, slot_id: UUID) -> AsyncGenerator[LiveStat]:
        async for payload in self.live_stats_model.listen_for_shots(slot_id):
            yield payload
