from uuid import UUID

from asyncpg import Pool
from fastapi import HTTPException, status

from models import SessionModel, ShotModel, SlotModel, TargetModel
from schema import SlotFilter, SlotRead


class BaseManager:
    """Base manager class that initializes common models and provides shared utilities."""

    def __init__(self, db_pool: Pool) -> None:
        self.session = SessionModel(db_pool)
        self.target = TargetModel(db_pool)
        self.slot = SlotModel(db_pool)
        self.shot = ShotModel(db_pool)

    def verify_archer_identity(
        self, current_archer_id: UUID, archer_id: UUID, detail: str = "Forbidden"
    ) -> None:
        """
        Verify that the authenticated archer is the same as the target archer.
        Used to enforce that archers can only mutate their own records.
        """
        if current_archer_id != archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=detail,
            )

    async def verify_slot_ownership(
        self, slot_id: UUID, current_archer_id: UUID, detail: str = "Forbidden"
    ) -> SlotRead:
        """
        Verify that the authenticated archer owns the given slot.
        Raises 403 Forbidden if they do not.
        """
        slot_row = await self.slot.get_one(SlotFilter(slot_id=slot_id))
        if current_archer_id != slot_row.archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=detail,
            )
        return slot_row
