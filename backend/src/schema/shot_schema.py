from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field, model_validator


class ShotBase(BaseModel):
    slot_id: UUID = Field(..., description="Slot identifier (UUID) this shot belongs to")
    x: float | None = Field(default=None, description="X coordinate in millimeters")
    y: float | None = Field(default=None, description="Y coordinate in millimeters")
    is_x: bool = Field(default=False, description="Indicates if the shot is an 'X' shot")
    score: int | None = Field(
        default=None,
        description="Shot score (0..10)",
        ge=0,
        le=10,
    )
    arrow_id: UUID | None = Field(
        default=None, description="Optional arrow identifier (e.g., arrow number or code)"
    )

    model_config = ConfigDict(title="Shot Base", extra="forbid")

    @model_validator(mode="after")
    def _validate_all_or_none(self) -> ShotBase:
        """Enforce that all three values (x, y, score) are either all NULL or all present."""

        if not (
            (self.x is None and self.y is None and self.score is None)
            or (self.x is not None and self.y is not None and self.score is not None)
        ):
            raise ValueError("x, y, and score must all be provided together or all be None")
        return self


class ShotCreate(ShotBase):
    model_config = ConfigDict(title="Shot Create", extra="forbid", populate_by_name=True)


class ShotSet(BaseModel):
    """This is just a placeholder. We don't want to allow updating shot fields."""

    model_config = ConfigDict(title="Shot Set", extra="forbid")


class ShotFilter(BaseModel):
    shot_id: UUID | None = Field(default=None, description="Shot identifier (UUID)")
    slot_id: UUID | None = Field(default=None, description="Filter by slot identifier (UUID)")
    x: float | None = Field(default=None, description="Filter by X coordinate in millimeters")
    y: float | None = Field(default=None, description="Filter by Y coordinate in millimeters")
    score: int | None = Field(
        default=None, description="Filter by shot score (non-negative integer)", ge=0, le=10
    )
    arrow_id: UUID | None = Field(
        default=None, description="Optional arrow identifier (e.g., arrow number or code)"
    )
    created_at: datetime | None = Field(
        default=None, description="Filter by creation timestamp (UTC)"
    )

    model_config = ConfigDict(title="Shot Filter", extra="forbid", populate_by_name=True)


class ShotUpdate(BaseModel):
    where: ShotFilter = Field(..., description="Filter criteria to select shots to update")
    data: ShotSet = Field(..., description="Fields to update on the selected shots")

    model_config = ConfigDict(title="Shot Update", extra="forbid")


class ShotRead(ShotBase):
    shot_id: UUID = Field(
        ...,
        description="Shot identifier (UUID)",
    )
    created_at: datetime = Field(
        ...,
        description="Creation timestamp (UTC)",
    )

    model_config = ConfigDict(title="Shot Read", extra="forbid", populate_by_name=True)

    def get_id(self) -> UUID:
        return self.shot_id


class ShotId(BaseModel):
    shot_id: UUID = Field(
        ...,
        description="Shot identifier (UUID)",
    )
    model_config = ConfigDict(title="Shot ID", extra="forbid", populate_by_name=True)


class ShotScore(BaseModel):
    shot_id: UUID = Field(..., description="Shot identifier (UUID)")
    score: int = Field(..., description="Shot score", ge=0, le=10)

    model_config = ConfigDict(title="Shot Score", extra="forbid")


class LiveStat(BaseModel):
    slot_id: UUID = Field(..., description="Slot identifier (UUID)")
    number_of_shots: int = Field(..., description="Total number of shots", ge=0)
    total_score: int = Field(..., description="Sum of all shot scores", ge=0)
    max_score: int = Field(..., description="Highest score achieved", ge=0, le=10)
    mean: float = Field(..., description="Average score")

    model_config = ConfigDict(title="Live Stat", extra="forbid")


class Stats(BaseModel):
    shot: ShotScore = Field(..., description="The latest shot score info")
    stats: LiveStat = Field(..., description="Aggregated live statistics")

    model_config = ConfigDict(title="Stats", extra="forbid")
