from datetime import UTC, datetime, timedelta
from enum import Enum, auto
from typing import Final
from uuid import UUID

from fastapi import HTTPException, status

from core.base_manager import BaseManager
from models.parent_model import DBNotFound
from schema import ShotCreate, ShotFilter, ShotId, ShotRead, SlotRead

MIN_BATCH_SIZE: Final[int] = 3
MAX_BATCH_SIZE: Final[int] = 10


class TimingScenario(Enum):
    NORMAL = auto()
    COMPRESSED = auto()  # Overlapping bounds (Scenario 2)
    DELAYED = auto()  # Way too slow (Scenario 3)


def _classify_scenario(available_time: float, default_duration: float) -> TimingScenario:
    if available_time < default_duration:
        return TimingScenario.COMPRESSED
    elif available_time > default_duration * 3:  # Threshold for "way slower"
        return TimingScenario.DELAYED
    return TimingScenario.NORMAL


def deal_with_delayed_scenario(
    window_start: datetime,
    window_end: datetime,
    shots: list[ShotCreate],
    latest_shot_time: datetime,
) -> datetime:
    return window_start


def deal_with_normal_scenario(
    window_start: datetime,
    window_end: datetime,
    shots: list[ShotCreate],
    latest_shot_time: datetime,
) -> datetime:
    return window_start


def deal_with_compress_scenario(
    window_start: datetime,
    window_end: datetime,
    shots: list[ShotCreate],
    latest_shot_time: datetime,
) -> datetime:
    # Scenario 2: Prevent overlap with previous shots.
    # Add a minimum physical gap (e.g. 1 second)
    min_start = latest_shot_time + timedelta(seconds=1)
    window_start = max(window_start, min_start)

    # Edge case: if window_start is somehow after window_end (e.g. latest_shot_time is in the future)
    if window_start >= window_end:
        window_start = window_end - timedelta(seconds=len(shots))
    return window_start


class ShotManagerError(Exception):
    """Custom exception for shot assignment manager errors."""


class ShotManager(BaseManager):
    async def _assign_dynamic_timestamps(self, shots: list[ShotCreate], slot: SlotRead) -> None:
        """Dynamically assigns created_at backward from now() for a batch of shots."""

        now_time = datetime.now(UTC)
        interval_seconds = getattr(slot, "interval_seconds", 20) or 20
        default_duration = interval_seconds * len(shots)

        # 1. Determine the absolute boundaries
        window_end = now_time
        window_start = now_time - timedelta(seconds=default_duration)

        # 2. Adjust boundaries based on constraints (The Scenarios)
        latest_shot_time = await self.shot.get_latest_shot_time(shots[0].slot_id)

        actions = {
            TimingScenario.COMPRESSED: deal_with_compress_scenario,
            TimingScenario.DELAYED: deal_with_delayed_scenario,
            TimingScenario.NORMAL: deal_with_normal_scenario,
        }
        if latest_shot_time is not None:
            available_time = (now_time - latest_shot_time).total_seconds()
            scenario = _classify_scenario(available_time, default_duration)

            window_start = actions[scenario](window_start, window_end, shots, latest_shot_time)

        # 3. Distribute evenly within the finalized window
        total_window_seconds = (window_end - window_start).total_seconds()
        actual_interval = total_window_seconds / (len(shots) - 1)

        for i, shot in enumerate(shots):
            shot.created_at = window_start + timedelta(seconds=actual_interval * i)

    async def create_single_shot(self, shot: ShotCreate, current_archer_id: UUID) -> UUID:
        # Verify that the slot belongs to the archer
        slot = await self.verify_slot_ownership(shot.slot_id, current_archer_id)

        # Verify session is open
        if not await self.session.does_open_session_exist(slot.session_id):
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
                detail="Cannot add shots to a closed session",
            )

        return await self.shot.insert_one(shot)

    async def create_batch_shots(
        self, shots: list[ShotCreate], current_archer_id: UUID
    ) -> list[UUID]:
        if not (MIN_BATCH_SIZE <= len(shots) <= MAX_BATCH_SIZE):
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid input")

        slot_ids = {shot.slot_id for shot in shots}
        if len(slot_ids) != 1:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="All shots must belong to the same slot",
            )

        slot_id = slot_ids.pop()
        slot = await self.verify_slot_ownership(slot_id, current_archer_id)

        # Verify session is open
        if not await self.session.does_open_session_exist(slot.session_id):
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
                detail="Cannot add shots to a closed session",
            )

        # --- Dynamic created_at calculation ---
        await self._assign_dynamic_timestamps(shots, slot)

        return await self.shot.insert_many(shots)

    async def create(
        self, shots: ShotCreate | list[ShotCreate], current_archer_id: UUID
    ) -> ShotId | list[ShotId]:
        try:
            if isinstance(shots, list):
                shot_ids = await self.create_batch_shots(shots, current_archer_id)
                return [ShotId(shot_id=shot_id) for shot_id in shot_ids]
            else:  # isinstance(shots, ShotCreate)
                shot_id = await self.create_single_shot(shots, current_archer_id)
                return ShotId(shot_id=shot_id)
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
        except TypeError:  # Original `else` block for invalid input type
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid input")

    async def get_shots_by_slot(self, slot_id: UUID, current_archer_id: UUID) -> list[ShotRead]:
        try:
            await self.verify_slot_ownership(slot_id, current_archer_id)
            return await self.shot.get_all(ShotFilter(slot_id=slot_id), [])
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e

    async def get_shots_count_by_slot(self, slot_id: UUID, current_archer_id: UUID) -> int:
        try:
            await self.verify_slot_ownership(slot_id, current_archer_id)
            return await self.shot.count_by_slot(slot_id)
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
