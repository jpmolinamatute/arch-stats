from uuid import UUID

from pydantic import BaseModel, Field, ConfigDict


class TargetsCreate(BaseModel):
    max_x_coordinate: float = Field(..., description="Max X coordinate of the target")
    max_y_coordinate: float = Field(..., description="Max Y coordinate of the target")
    radius: list[float] = Field(..., description="List of radii for target rings")
    points: list[int] = Field(
        ...,
        description="List of points for each ring (must match radii length)",
    )
    height: float = Field(..., description="Height of the target")
    human_identifier: str = Field(..., max_length=10, description="Human-friendly identifier")
    session_id: UUID = Field(..., description="ID of the session this target belongs to")


class TargetsUpdate(BaseModel):
    max_x_coordinate: float | None = Field(None, description="Max X coordinate of the target")
    max_y_coordinate: float | None = Field(None, description="Max Y coordinate of the target")
    radius: list[float] | None = Field(None, description="List of radii for target rings")
    points: list[int] | None = Field(None, description="List of points for each ring")
    height: float | None = Field(None, description="Height of the target")
    human_identifier: str | None = Field(None, description="Optional human-friendly identifier")
    session_id: UUID | None = Field(None, description="ID of the session this target belongs to")


class TargetsRead(BaseModel):
    target_id: UUID = Field(..., alias="id", description="Unique target ID")
    max_x_coordinate: float = Field(..., description="Max X coordinate of the target")
    max_y_coordinate: float = Field(..., description="Max Y coordinate of the target")
    radius: list[float] = Field(..., description="List of radii for target rings")
    points: list[int] = Field(
        ..., description="List of points for each ring (matches radii length)"
    )
    height: float = Field(..., description="Height of the target")
    human_identifier: str = Field(..., description="Optional human-friendly identifier")
    session_id: UUID = Field(..., description="ID of the session this target belongs to")

    model_config = ConfigDict(populate_by_name=True, extra="forbid")
