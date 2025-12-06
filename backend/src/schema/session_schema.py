from __future__ import annotations

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field, field_validator


class SessionCreate(BaseModel):
    owner_archer_id: UUID = Field(description="Owner archer UUID")
    session_location: str = Field(
        min_length=1, max_length=255, description="Session location label (e.g., range name)"
    )
    is_indoor: bool = Field(default=False, description="Wether the session is indoors or not")
    is_opened: bool = Field(..., description="Wether the session is open or not")
    shot_per_round: int = Field(
        default=6,
        ge=3,
        le=10,
        description="Number of shots per round for the current session format",
    )
    model_config = ConfigDict(title="Session Create", extra="forbid", populate_by_name=True)


class SessionSet(BaseModel):
    closed_at: datetime | None = Field(default=None, description="Closing timestamp (UTC)")
    session_location: str | None = Field(
        default=None, min_length=1, max_length=255, description="New session location label"
    )
    is_opened: bool | None = Field(default=None, description="Open/close session toggle")
    is_indoor: bool | None = Field(default=None, description="Set indoor flag")
    shot_per_round: int | None = Field(
        default=None,
        ge=3,
        le=10,
        description="Number of shots per round for the current session format",
    )
    model_config = ConfigDict(title="Session Set", extra="forbid")


class SessionFilter(BaseModel):
    session_id: UUID | None = Field(default=None, description="Session identifier (UUID)")
    owner_archer_id: UUID | None = Field(default=None, description="Filter by owner archer UUID")
    created_at: datetime | None = Field(
        default=None,
        description="Open timestamp (UTC)",
    )
    closed_at: datetime | None = Field(default=None, description="Closing timestamp (UTC)")
    shot_per_round: int | None = Field(
        default=None,
        ge=3,
        le=10,
        description="Number of shots per round for the current session format",
    )
    session_location: str | None = Field(
        default=None,
        min_length=1,
        max_length=255,
        description="Session location label (e.g., range name)",
    )
    is_opened: bool | None = Field(default=None, description="Filter by open sessions")
    is_indoor: bool | None = Field(default=None, description="Filter by indoor/outdoor")

    model_config = ConfigDict(title="Session Filter", extra="forbid", populate_by_name=True)


class SessionUpdate(BaseModel):
    where: SessionFilter = Field(..., description="Filter criteria to select sessions to update")
    data: SessionSet = Field(..., description="Fields to update on the selected sessions")

    model_config = ConfigDict(title="Session Update", extra="forbid")

    @field_validator("data")
    @classmethod
    def _validate_data_not_empty(cls, v: SessionSet) -> SessionSet:
        if len(v.model_fields_set) == 0:
            raise ValueError("data must set at least one field")
        return v

    @field_validator("where")
    @classmethod
    def _validate_where_has_id(cls, v: SessionFilter) -> SessionFilter:
        if len(v.model_fields_set) == 0:
            raise ValueError("where must set at least one field")
        elif v.session_id is None:
            raise ValueError("where.session_id must be provided")
        return v


class SessionRead(SessionCreate):
    session_id: UUID = Field(
        ...,
        description="Session identifier (UUID)",
    )
    created_at: datetime = Field(
        default=...,
        description="Open timestamp (UTC)",
    )
    closed_at: datetime | None = Field(default=None, description="Closing timestamp (UTC), if any")

    model_config = ConfigDict(title="Session Read", extra="forbid", populate_by_name=False)

    def get_id(self) -> UUID:
        return self.session_id


class SessionId(BaseModel):
    session_id: UUID | None = Field(
        default=None,
        description="Session identifier (UUID)",
    )
    model_config = ConfigDict(title="Session ID", extra="forbid", populate_by_name=True)
