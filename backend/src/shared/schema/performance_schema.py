from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class PerformanceRead(BaseModel):
    """
    Read model for the performance SQL view.
    Represents a shot enriched with target context and computed metrics.
    """

    shot_id: UUID = Field(..., alias="id", description="Shot identifier (row id from shots)")
    session_id: UUID = Field(..., description="Session this shot belongs to")
    arrow_id: UUID = Field(..., description="Arrow used for this shot")

    arrow_engage_time: datetime = Field(..., description="When the arrow engaged the bow")
    arrow_disengage_time: datetime = Field(..., description="When the arrow left the bow")
    arrow_landing_time: datetime | None = Field(
        default=None, description="When the arrow landed on target (None on miss)"
    )

    x: float | None = Field(default=None, description="Landing X coordinate (None on miss)")
    y: float | None = Field(default=None, description="Landing Y coordinate (None on miss)")

    time_of_flight_seconds: float | None = Field(
        default=None, description="Flight time in seconds (None if no landing time)"
    )
    arrow_speed: float | None = Field(
        default=None, description="Estimated arrow speed in m/s (None if flight time unavailable)"
    )
    score: int | None = Field(
        default=None,
        description="Points scored (NULL if no faces configured; 0 for miss; >0 for hit)",
    )

    # Arrow fields (flattened from arrows table)
    length: float = Field(..., description="Arrow's total length in cm")
    human_identifier: str = Field(..., max_length=10, description="Arrow short unique identifier")
    is_programmed: bool = Field(..., description="Whether the arrow has been programmed")
    registration_date: datetime = Field(..., description="Arrow registration datetime")
    voided_date: datetime | None = Field(
        default=None, description="When the arrow was voided (no longer in use)"
    )
    is_active: bool = Field(..., description="Whether the arrow is active")
    label_position: float | None = Field(
        default=None, description="Position of label from nock in cm"
    )
    weight: float | None = Field(default=None, description="Arrow's weight in grams")
    diameter: float | None = Field(default=None, description="Diameter of the arrow in mm")
    spine: float | None = Field(default=None, description="Arrow spine (flexibility rating)")

    model_config = ConfigDict(extra="forbid")

    def get_id(self) -> UUID:
        return self.shot_id


class PerformanceFilter(BaseModel):
    """
    Filter model for the performance SQL view.
    Represents a shot enriched with target context and computed metrics.
    """

    shot_id: UUID | None = Field(
        default=None, alias="id", description="Shot identifier (row id from shots)"
    )
    session_id: UUID | None = Field(default=None, description="Session this shot belongs to")
    arrow_id: UUID | None = Field(default=None, description="Arrow used for this shot")

    arrow_engage_time: datetime | None = Field(
        default=None, description="When the arrow engaged the bow"
    )
    arrow_disengage_time: datetime | None = Field(
        default=None, description="When the arrow left the bow"
    )
    arrow_landing_time: datetime | None = Field(
        default=None, description="When the arrow landed on target (None on miss)"
    )

    x: float | None = Field(default=None, description="Landing X coordinate (None on miss)")
    y: float | None = Field(default=None, description="Landing Y coordinate (None on miss)")

    time_of_flight_seconds: float | None = Field(
        default=None, description="Flight time in seconds (None if no landing time)"
    )
    arrow_speed: float | None = Field(
        default=None, description="Estimated arrow speed in m/s (None if flight time unavailable)"
    )
    score: int | None = Field(
        default=None,
        description="Points scored (NULL if no faces configured; 0 for miss; >0 for hit)",
    )

    # Arrow fields (flattened from arrows table)
    length: float | None = Field(default=None, description="Arrow's total length in cm")
    human_identifier: str | None = Field(
        default=None, max_length=10, description="Arrow short unique identifier"
    )
    is_programmed: bool | None = Field(
        default=None, description="Whether the arrow has been programmed"
    )
    registration_date: datetime | None = Field(
        default=None, description="Arrow registration datetime"
    )
    voided_date: datetime | None = Field(
        default=None, description="When the arrow was voided (no longer in use)"
    )
    is_active: bool | None = Field(default=None, description="Whether the arrow is active")
    label_position: float | None = Field(
        default=None, description="Position of label from nock in cm"
    )
    weight: float | None = Field(default=None, description="Arrow's weight in grams")
    diameter: float | None = Field(default=None, description="Diameter of the arrow in mm")
    spine: float | None = Field(default=None, description="Arrow spine (flexibility rating)")

    model_config = ConfigDict(extra="forbid")
