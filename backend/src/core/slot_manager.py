from uuid import UUID

from fastapi import HTTPException, status

from core.base_manager import BaseManager
from models.parent_model import DBException, DBNotFound
from schema import (
    BowStyleType,
    FaceType,
    SlotFilter,
    SlotJoinRequest,
    SlotJoinResponse,
    SlotLeaveRequest,
    SlotLetterType,
    SlotSet,
)
from schema.slot_schema import FullSlotInfo


class SlotManagerError(Exception):
    """Custom exception for slot assignment manager errors."""


class SlotManager(BaseManager):
    async def create_target(self, session_id: UUID, distance: int, lane: int) -> UUID:
        return await self.target.create_one(
            session_id=session_id,
            distance=distance,
            lane=lane,
        )

    async def create_slotassignment(  # noqa: PLR0913
        self,
        *,
        session_id: UUID,
        target_id: UUID,
        archer_id: UUID,
        face_type: FaceType,
        slot_letter: SlotLetterType,
        bowstyle: BowStyleType,
        draw_weight: float,
        club_id: UUID | None,
        shot_per_round: int | None,
        interval_seconds: int,
    ) -> UUID:
        new_slot_id = await self.slot.create_one(
            session_id=session_id,
            target_id=target_id,
            archer_id=archer_id,
            face_type=face_type,
            slot_letter=slot_letter,
            bowstyle=bowstyle,
            draw_weight=draw_weight,
            club_id=club_id,
            shot_per_round=shot_per_round,
            interval_seconds=interval_seconds,
        )
        await self.slot.refresh_open_participants()
        return new_slot_id

    async def assign_archer_to_slot(self, req_data: SlotJoinRequest) -> SlotJoinResponse:
        """Assigns an archer to a target slot within a session."""
        try:
            exists = await self.session.does_open_session_exist(req_data.session_id)

            if not exists:
                raise DBNotFound("ERROR: Session either doesn't exist or it was already closed")

            # Prevent duplicate participation
            current_participation = await self.session.is_archer_participating(req_data.archer_id)
            if current_participation is not None:
                if current_participation == req_data.session_id:
                    raise SlotManagerError("ERROR: archer already joined this session")
                raise SlotManagerError("ERROR: archer already participating in an open session")

            available_targets = await self.slot.get_available_targets(req_data)
            lane = 1
            if available_targets:
                target_id = available_targets[0].get_id()
                lane = available_targets[0].lane
                used_letters = await self.slot.get_assigned_letters(target_id)
                for opt in [
                    SlotLetterType.A,
                    SlotLetterType.B,
                    SlotLetterType.C,
                    SlotLetterType.D,
                ]:
                    if opt not in used_letters:
                        letter = opt
                        break
                else:
                    raise SlotManagerError("ERROR: selected target is full")
            else:
                lane = await self.target.get_next_empty_lane(req_data.session_id)
                target_id = await self.create_target(
                    session_id=req_data.session_id, distance=req_data.distance, lane=lane
                )
                letter = SlotLetterType.A

            slot_id = await self.create_slotassignment(
                session_id=req_data.session_id,
                target_id=target_id,
                archer_id=req_data.archer_id,
                face_type=req_data.face_type,
                slot_letter=letter,
                bowstyle=req_data.bowstyle,
                draw_weight=req_data.draw_weight,
                club_id=req_data.club_id,
                shot_per_round=req_data.shot_per_round,
                interval_seconds=req_data.interval_seconds,
            )
            return SlotJoinResponse(slot_id=slot_id, slot=f"{lane}{letter.value}")
        except DBNotFound as e:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)
            ) from e
        except SlotManagerError as e:
            msg = str(e)
            if msg in (
                "ERROR: archer already participating in an open session",
                "ERROR: archer already joined this session",
            ):
                raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=msg) from e
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=msg) from e

    async def re_join_session(self, slot_id: UUID, current_archer_id: UUID) -> SlotJoinResponse:
        """Re-activate a previously inactive slot assignment."""
        try:
            slot_row = await self.verify_slot_ownership(
                slot_id, current_archer_id, detail="ERROR: user not allowed to re-join"
            )

            if slot_row.is_shooting:
                raise DBNotFound(
                    "ERROR: the archer is either not allowed to re-join or they are already in"
                )

            exists = await self.session.does_open_session_exist(slot_row.session_id)
            if not exists:
                raise DBNotFound("ERROR: The session doesn't exist or it was already closed")

            where = SlotFilter(slot_id=slot_id)
            await self.slot.update(SlotSet(is_shooting=True), where)
            await self.slot.refresh_open_participants()
            slot_row = await self.slot.get_slot_with_lane(slot_id)
            if slot_row.slot is None:
                raise SlotManagerError("ERROR: slot is unexpectedly None")
            return SlotJoinResponse(slot_id=slot_row.slot_id, slot=slot_row.slot)
        except DBNotFound as e:
            msg = str(e)
            if msg == "ERROR: The session doesn't exist or it was already closed":
                raise HTTPException(
                    status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
                ) from e
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
                detail="ERROR: the archer is either not allowed to re-join or they are already in",
            ) from e
        except DBException as e:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
        except SlotManagerError as e:
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e)) from e

    async def leave_session(self, slot_id: UUID, current_archer_id: UUID) -> None:
        """Deactivate an active slot assignment (leave the session)."""
        try:
            slot_row = await self.verify_slot_ownership(
                slot_id, current_archer_id, detail="ERROR: user not allowed to leave"
            )

            if not slot_row.is_shooting:
                raise DBNotFound("ERROR: archer is not participating in this session")

            exists = await self.session.does_open_session_exist(slot_row.session_id)
            if not exists:
                raise DBNotFound("ERROR: Session either doesn't exist or it was already closed")

            req = SlotLeaveRequest(session_id=slot_row.session_id, archer_id=slot_row.archer_id)
            await self.slot.leave_session(req)
        except DBNotFound as e:
            msg = str(e)
            if msg == "ERROR: Session either doesn't exist or it was already closed":
                raise HTTPException(
                    status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
                ) from e
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail="ERROR: archer is not participating in this session",
            ) from e

    async def get_full_slot_info(
        self, current_archer_id: UUID, slot_id: UUID | None = None, archer_id: UUID | None = None
    ) -> FullSlotInfo:
        try:
            if archer_id is not None and current_archer_id != archer_id:
                raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

            info = await self.slot.get_full_slot_info(slot_id=slot_id, archer_id=archer_id)

            if slot_id is not None and current_archer_id != info.archer_id:
                raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

            return info
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
