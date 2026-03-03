from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field

from schema.base import BaseUpdateValidation


class TargetRead(BaseModel):
    target_id: UUID = Field(
        ...,
        description="Unique identifier for the target",
    )
    created_at: datetime = Field(default=..., description="Creation timestamp (UTC)")
    session_id: UUID = Field(..., description="Session UUID this target belongs to")
    distance: int = Field(..., description="Target distance in meters")
    lane: int = Field(..., description="Lane number for the target")
    occupied: int | None = Field(None, description="Number of occupied slots for the target")

    model_config = ConfigDict(title="Target Read", extra="forbid", populate_by_name=False)

    def get_id(self) -> UUID:
        return self.target_id


class TargetCreate(BaseModel):
    session_id: UUID = Field(..., description="Session UUID this target belongs to")
    distance: int = Field(..., ge=1, le=100, description="Target distance in meters (1-100)")
    lane: int = Field(..., description="Lane number for the target")

    model_config = ConfigDict(title="Target Create", extra="forbid")


class TargetSet(BaseModel):
    distance: int | None = Field(
        None, ge=1, le=100, description="Update the target distance in meters (1-100)"
    )
    lane: int | None = Field(None, description="Update the lane number for the target")

    model_config = ConfigDict(title="Target Set", extra="forbid")


class TargetFilter(BaseModel):
    target_id: UUID | None = Field(default=None, description="Unique identifier for the target")
    created_at: datetime | None = Field(default=None, description="Creation timestamp (UTC)")
    session_id: UUID | None = Field(None, description="Filter by session UUID")
    distance: int | None = Field(None, description="Filter by target distance in meters")
    lane: int | None = Field(None, description="Filter by lane number")

    model_config = ConfigDict(title="Target Filter", extra="forbid")


class TargetUpdate(BaseUpdateValidation):
    _id_field_name = "target_id"
    where: TargetFilter = Field(..., description="Filter criteria to select targets to update")
    data: TargetSet = Field(..., description="Fields to update on the selected targets")

    model_config = ConfigDict(title="Target Update", extra="forbid")
