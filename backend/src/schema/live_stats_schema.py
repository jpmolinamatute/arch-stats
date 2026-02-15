from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class ShotScore(BaseModel):
    shot_id: UUID = Field(..., description="Shot identifier (UUID)")
    score: int = Field(..., description="Shot score", ge=0, le=10)
    is_x: bool = Field(default=False, description="Is X")
    created_at: datetime = Field(..., description="Creation timestamp")

    model_config = ConfigDict(title="Shot Score", extra="forbid")


class Stats(BaseModel):
    slot_id: UUID = Field(..., description="Slot identifier (UUID)")
    number_of_shots: int = Field(..., description="Total number of shots", ge=0)
    total_score: int = Field(..., description="Sum of all shot scores", ge=0)
    max_score: int = Field(..., description="Maximum possible score (number of shots * 10)", ge=0)
    mean: float = Field(..., description="Average score")

    model_config = ConfigDict(title="Stats", extra="forbid")


class LiveStat(BaseModel):
    scores: list[ShotScore] = Field(..., description="List of latest shot score info")
    stats: Stats = Field(..., description="Aggregated live statistics")

    model_config = ConfigDict(title="Live Stat", extra="forbid")
