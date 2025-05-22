from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class ShotsCreate(BaseModel):
    arrow_id: UUID = Field(..., description="ID of the arrow used for this shot")
    arrow_engage_time: datetime = Field(..., description="Timestamp when arrow was engaged")
    arrow_disengage_time: datetime = Field(..., description="Timestamp when arrow was disengaged")
    arrow_landing_time: datetime | None = Field(
        None, description="Timestamp when arrow landed (optional for missed shots)"
    )
    x_coordinate: float | None = Field(None, description="X coordinate of the arrow on the target")
    y_coordinate: float | None = Field(None, description="Y coordinate of the arrow on the target")


class ShotsUpdate(BaseModel):
    arrow_id: UUID | None = Field(None, description="ID of the arrow used for this shot")
    arrow_engage_time: datetime | None = Field(None, description="Timestamp when arrow was engaged")
    arrow_disengage_time: datetime | None = Field(
        None, description="Timestamp when arrow was disengaged"
    )
    arrow_landing_time: datetime | None = Field(
        None, description="Timestamp when arrow landed (optional for missed shots)"
    )
    x_coordinate: float | None = Field(None, description="X coordinate of the arrow on the target")
    y_coordinate: float | None = Field(None, description="Y coordinate of the arrow on the target")


class ShotsRead(BaseModel):
    shot_id: UUID = Field(..., alias="id", description="Unique shot ID")
    arrow_id: UUID = Field(..., description="ID of the arrow used for this shot")
    arrow_engage_time: datetime = Field(..., description="Timestamp when arrow was engaged")
    arrow_disengage_time: datetime = Field(..., description="Timestamp when arrow was disengaged")
    arrow_landing_time: datetime | None = Field(
        None, description="Timestamp when arrow landed (optional for missed shots)"
    )
    x_coordinate: float | None = Field(None, description="X coordinate of the arrow on the target")
    y_coordinate: float | None = Field(None, description="Y coordinate of the arrow on the target")

    model_config = ConfigDict(populate_by_name=True, extra="forbid")
