from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class ShotsCreate(BaseModel):
    model_config = ConfigDict(extra="forbid")


class ShotsUpdate(BaseModel):
    model_config = ConfigDict(extra="forbid")


class ShotsFilters(BaseModel):
    arrow_id: UUID | None = Field(None, description="ID of the arrow recorded")
    arrow_engage_time: datetime | None = Field(None, description="Filter by session start time")
    arrow_disengage_time: datetime | None = Field(None, description="Filter by session start time")
    arrow_landing_time: datetime | None = Field(None, description="Filter by session start time")
    x_coordinate: float | None = Field(None, description="X coordinate of the target")
    y_coordinate: float | None = Field(None, description="Y coordinate of the target")
    session_id: UUID | None = Field(None, description="ID of the session this shot belongs to")
    model_config = ConfigDict(extra="forbid")
