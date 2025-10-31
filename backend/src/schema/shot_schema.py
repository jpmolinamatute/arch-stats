from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field, ValidationInfo, field_validator


class ShotBase(BaseModel):
    slot_id: UUID = Field(..., description="Slot identifier (UUID) this shot belongs to")
    x: float | None = Field(default=None, description="X coordinate in millimeters")
    y: float | None = Field(default=None, description="Y coordinate in millimeters")
    score: int | None = Field(default=None, description="Shot score (non-negative integer)", ge=0)

    model_config = ConfigDict(title="Shot Base", extra="forbid")

    @field_validator("score", "x", "y")
    @classmethod
    def _validate_all_or_none(
        cls, v: float | int | None, info: ValidationInfo
    ) -> float | int | None:
        """Enforce that all three values (x, y, score) are either all NULL or all present."""
        values = info.data
        coords_score_fields = {"x", "y", "score"}
        set_fields = {k for k in coords_score_fields if k in values and values[k] is not None}

        # If this field is being set and any fields exist in data
        if len(set_fields) > 0 and len(set_fields) != len(coords_score_fields):
            raise ValueError("x, y, and score must all be provided together or all be None")
        return v


class ShotCreate(ShotBase):
    model_config = ConfigDict(title="Shot Create", extra="forbid", populate_by_name=True)


class ShotSet(BaseModel):
    x: float | None = Field(default=None, description="Updated X coordinate in millimeters")
    y: float | None = Field(default=None, description="Updated Y coordinate in millimeters")
    score: int | None = Field(
        default=None, description="Updated shot score (non-negative integer)", ge=0, le=10
    )

    model_config = ConfigDict(title="Shot Set", extra="forbid")


class ShotFilter(BaseModel):
    shot_id: UUID | None = Field(default=None, description="Shot identifier (UUID)")
    slot_id: UUID | None = Field(default=None, description="Filter by slot identifier (UUID)")
    x: float | None = Field(default=None, description="Filter by X coordinate in millimeters")
    y: float | None = Field(default=None, description="Filter by Y coordinate in millimeters")
    score: int | None = Field(
        default=None, description="Filter by shot score (non-negative integer)", ge=0, le=10
    )
    created_at: datetime | None = Field(
        default=None, description="Filter by creation timestamp (UTC)"
    )

    model_config = ConfigDict(title="Shot Filter", extra="forbid", populate_by_name=True)


class ShotUpdate(BaseModel):
    where: ShotFilter = Field(..., description="Filter criteria to select shots to update")
    data: ShotSet = Field(..., description="Fields to update on the selected shots")

    model_config = ConfigDict(title="Shot Update", extra="forbid")

    @field_validator("data")
    @classmethod
    def _validate_data_not_empty(cls, v: ShotSet) -> ShotSet:
        if len(v.model_fields_set) == 0:
            raise ValueError("data must set at least one field")
        return v

    @field_validator("where")
    @classmethod
    def _validate_where_has_id(cls, v: ShotFilter) -> ShotFilter:
        if len(v.model_fields_set) == 0:
            raise ValueError("where must set at least one field")
        elif v.shot_id is None:
            raise ValueError("where.shot_id must be provided")
        return v


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
