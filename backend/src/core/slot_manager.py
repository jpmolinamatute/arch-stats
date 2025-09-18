from uuid import UUID

from asyncpg import Pool

from models import SessionModel, SlotModel, TargetModel
from models.parent_model import DBNotFound
from schema import (
    BowStyleType,
    SlotFilter,
    SlotId,
    SlotJoinRequest,
    SlotJoinResponse,
    SlotLetterType,
    SlotSet,
    TargetFaceType,
)


class SlotManagerError(Exception):
    """Custom exception for slot assignment manager errors."""


class SlotManager:
    def __init__(self, db_pool: Pool) -> None:
        self.session = SessionModel(db_pool)
        self.target = TargetModel(db_pool)
        self.slot = SlotModel(db_pool)

    async def create_target(self, session_id: UUID, distance: int, lane: int) -> UUID:
        return await self.target.create_one(
            session_id=session_id,
            distance=distance,
            lane=lane,
        )

    # pylint: disable=[too-many-arguments]
    async def create_slotassignment(
        self,
        *,
        session_id: UUID,
        target_id: UUID,
        archer_id: UUID,
        face_type: TargetFaceType,
        slot_letter: SlotLetterType,
        bowstyle: BowStyleType,
        draw_weight: float,
        club_id: UUID | None,
    ) -> UUID:
        return await self.slot.create_one(
            session_id=session_id,
            target_id=target_id,
            archer_id=archer_id,
            face_type=face_type,
            slot_letter=slot_letter,
            bowstyle=bowstyle,
            draw_weight=draw_weight,
            club_id=club_id,
        )

    async def assign_archer_to_slot(self, req_data: SlotJoinRequest) -> SlotId:
        """Assigns an archer to a target slot within a session.

        This method implements the business logic for assigning an archer to a target slot
        according to the session flow rules (see session_flow.md). It finds or creates a suitable
        target for the archer's distance, assigns the next available slot (A-D), and creates the
        corresponding slot assignment record. Returns the assignment details.
        """

        exists = await self.session.does_session_exist(req_data.session_id)

        if not exists:
            raise DBNotFound("ERROR: Session either doesn't exist or it was already closed")

        # Prevent duplicate participation
        current_participation = await self.session.is_archer_participating(req_data.archer_id)
        if current_participation is not None:
            # If archer is already participating in this same session, use a clearer message
            if current_participation == req_data.session_id:
                raise SlotManagerError("ERROR: archer already joined this session")
            # Otherwise, they are participating in a different open session
            raise SlotManagerError("ERROR: archer already participating in an open session")

        available_targets = await self.slot.get_available_targets(req_data)
        lane = 1
        if available_targets:
            target_id = available_targets[0].get_id()
            lane = available_targets[0].lane
            # Determine next free letter based on currently shooting assignments
            used_letters = await self.slot.get_assigned_letters(target_id)
            for opt in [SlotLetterType.A, SlotLetterType.B, SlotLetterType.C, SlotLetterType.D]:
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
        )
        return SlotId(slot_id=slot_id)

    async def re_join_session(
        self, session_req: SlotJoinRequest, current_archer_id: UUID
    ) -> SlotJoinResponse:
        # Find an existing inactive assignment for this archer (optionally scoped by session)

        if session_req.archer_id != current_archer_id:
            raise SlotManagerError("ERROR: user not allowed to re-join")

        where = SlotFilter(
            is_shooting=False,
            archer_id=session_req.archer_id,
            session_id=session_req.session_id,
        )
        slot_row = await self.slot.get_one(where)

        # Reactivate the assignment
        await self.slot.update(SlotSet(is_shooting=True), where)

        # Fetch lane from target to compose slot code (e.g., "1A")
        lane = await self.target.get_lane(slot_row.target_id)

        return SlotJoinResponse(
            slot_id=slot_row.slot_id,
            target_id=slot_row.target_id,
            slot=f"{lane}{slot_row.slot_letter.value}",
        )
