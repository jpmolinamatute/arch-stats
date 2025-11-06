from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field, field_validator

from schema.enums import BowStyleType, SlotLetterType
from schema.face_schema import FaceType


# SlotCreate and SlotJoinRequest are involved in the creation of a row in the
# slots table. However, since the client is going to "Request" for a slot availability,
# the client doesn't know the target_id nor the slot. In contrast, SlotCreate doesn't care
# about distance and expects a target_id and slot_letter


class SlotCommons(BaseModel):
    archer_id: UUID = Field(..., description="ID of the archer to assign to the slot")
    session_id: UUID = Field(..., description="ID of the session for the assignment")
    face_type: FaceType = Field(..., description="Type of target face (e.g., '40cm', '80cm')")
    is_shooting: bool = Field(
        default=True, description="Set to True if archer is currently shooting, False otherwise"
    )
    bowstyle: BowStyleType = Field(..., description="Archer bowstyle copied at join time")
    draw_weight: float = Field(..., description="Archer draw weight copied at join time")
    club_id: UUID | None = Field(default=None, description="Archer current club at join time")


class SlotCreate(SlotCommons):
    slot_letter: SlotLetterType = Field(..., description="Slot letter to assign (A-D)")
    target_id: UUID = Field(..., description="ID of the target to assign the slot to")
    model_config = ConfigDict(title="Slot Assignment Create", extra="forbid")


class SlotJoinRequest(SlotCommons):
    distance: int = Field(
        ..., ge=1, le=100, description="Requested distance to shoot (meters 1-100)"
    )

    model_config = ConfigDict(title="Slot Assignment Request", extra="forbid")


class SlotReJoinRequest(BaseModel):
    slot_id: UUID = Field(..., description="ID of the slot assignment to re-join")
    session_id: UUID = Field(..., description="ID of the session to re-join")
    archer_id: UUID = Field(..., description="ID of the archer re-joining the session")
    model_config = ConfigDict(title="Slot Re-Join Request", extra="forbid")


class SlotJoinResponse(BaseModel):
    slot_id: UUID = Field(..., description="ID of the slot assignment")
    slot: str = Field(..., description="Slot assigned to the archer (1A, 3D, 4B)")

    model_config = ConfigDict(title="Slot Assignment Response", extra="forbid")


class SlotLeaveRequest(BaseModel):
    archer_id: UUID = Field(..., description="Archer leaving the session")
    session_id: UUID = Field(..., description="ID of the session to leave")

    model_config = ConfigDict(title="Leave Session Request", extra="forbid")


class SlotId(BaseModel):
    slot_id: UUID = Field(..., description="Unique identifier for the slot assignment")

    model_config = ConfigDict(title="Slot Assignment ID", extra="forbid")


class SlotRead(BaseModel):
    slot_id: UUID = Field(
        ...,
        description="Unique identifier for the slot assignment",
    )
    target_id: UUID = Field(..., description="ID of the target this slot is assigned to")
    archer_id: UUID = Field(..., description="ID of the archer assigned to this slot")
    session_id: UUID = Field(..., description="ID of the session this assignment belongs to")
    face_type: FaceType = Field(..., description="Type of target face (e.g., '40cm', '80cm')")
    slot_letter: SlotLetterType = Field(..., description="Slot letter assigned to the archer (A-D)")
    slot: str | None = Field(
        default=None, description="Slot code combining lane and letter (e.g., '1A')"
    )
    created_at: datetime | None = Field(
        default=None,
        description="Assignment timestamp (UTC)",
    )
    is_shooting: bool = Field(
        ..., description="Whether the archer is currently shooting in this slot"
    )
    bowstyle: BowStyleType = Field(..., description="Archer bowstyle copied at join time")
    draw_weight: float = Field(
        ..., description="Archer draw weight copied at join time", gt=0, le=200
    )
    club_id: UUID | None = Field(default=None, description="Archer current club at join time")

    model_config = ConfigDict(title="Slot Assignment Read", extra="forbid", populate_by_name=False)

    def get_id(self) -> UUID:
        return self.slot_id


class SlotSet(BaseModel):
    is_shooting: bool | None = Field(
        None, description="Set to True if archer is currently shooting, False otherwise"
    )
    face_type: FaceType | None = Field(
        None, description="Update the type of target face (e.g., '40cm', '80cm')"
    )
    slot_letter: SlotLetterType | None = Field(
        None, description="Update the slot letter assignment (A-D)"
    )

    model_config = ConfigDict(title="Slot Assignment Set", extra="forbid")


class SlotFilter(BaseModel):
    slot_id: UUID | None = Field(
        default=None, description="Unique identifier for the slot assignment"
    )
    target_id: UUID | None = Field(None, description="Filter by target ID")
    archer_id: UUID | None = Field(None, description="Filter by archer ID")
    session_id: UUID | None = Field(None, description="Filter by session ID")
    slot_letter: SlotLetterType | None = Field(
        None, description="Update the slot letter assignment (A-D)"
    )
    is_shooting: bool | None = Field(
        None, description="Filter by shooting status (True if currently shooting)"
    )
    created_at: datetime | None = Field(
        default=None,
        description="Assignment timestamp (UTC)",
    )
    model_config = ConfigDict(title="Slot Assignment Filter", extra="forbid")


class SlotUpdate(BaseModel):
    where: SlotFilter = Field(
        ..., description="Filter criteria to select slot assignments to update"
    )
    data: SlotSet = Field(..., description="Fields to update on the selected slot assignments")

    model_config = ConfigDict(title="Slot Assignment Update", extra="forbid")

    @field_validator("data")
    @classmethod
    def _validate_data_not_empty(cls, v: SlotSet) -> SlotSet:
        if len(v.model_fields_set) == 0:
            raise ValueError("data must set at least one field")
        return v

    @field_validator("where")
    @classmethod
    def _validate_where_has_id(cls, v: SlotFilter) -> SlotFilter:
        if len(v.model_fields_set) == 0:
            raise ValueError("where must set at least one field")
        elif v.slot_id is None:
            raise ValueError("where.slot_id must be provided")
        return v


class FullSlotInfo(BaseModel):
    slot_id: UUID = Field(..., description="Unique identifier for the slot assignment")
    target_id: UUID = Field(..., description="ID of the target this slot is assigned to")
    archer_id: UUID = Field(..., description="ID of the archer assigned to this slot")
    session_id: UUID = Field(..., description="ID of the session this assignment belongs to")
    face_type: FaceType = Field(..., description="Type of target face (e.g., '40cm', '80cm')")
    slot_letter: SlotLetterType = Field(..., description="Slot letter assigned to the archer (A-D)")
    is_shooting: bool = Field(
        ..., description="Whether the archer is currently shooting in this slot"
    )
    created_at: datetime = Field(..., description="Assignment timestamp (UTC)")
    lane: int = Field(..., description="Lane number of the target")
    distance: int = Field(..., ge=1, le=100, description="Target distance (meters 1-100)")
    slot: str = Field(..., description="Slot code combining lane and letter (e.g., '1A')")
    bowstyle: BowStyleType = Field(..., description="Archer bowstyle copied at join time")
    draw_weight: float = Field(
        ..., description="Archer draw weight copied at join time", gt=0, le=200
    )
    club_id: UUID | None = Field(default=None, description="Archer current club at join time")

    model_config = ConfigDict(title="Full Slot Assignment Info", extra="forbid")
