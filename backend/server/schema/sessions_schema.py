from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field


class SessionsCreate(BaseModel):
    is_opened: bool = Field(..., description="Update open/closed state")
    start_time: datetime = Field(..., description="Session start time (set by the app)")
    location: str = Field(..., max_length=255, description="Location of the session")
    model_config = ConfigDict(extra="forbid")


class SessionsUpdate(BaseModel):
    is_opened: bool | None = Field(None, description="Update open/closed state")
    location: str | None = Field(None, max_length=255, description="Location of the session")
    start_time: datetime | None = Field(None, description="Session start time")
    end_time: datetime | None = Field(None, description="Session end time")
    model_config = ConfigDict(extra="forbid")
