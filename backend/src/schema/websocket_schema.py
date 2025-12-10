from datetime import UTC, datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

from schema.shot_schema import Stats


class WebSocketMessage(BaseModel):
    ts: datetime = Field(
        default_factory=lambda: datetime.now(UTC),
        description="Timestamp of the event",
    )
    data_type: Literal["shot.created"] = "shot.created"
    data_content: Stats
    model_config = ConfigDict(populate_by_name=True, extra="forbid")
