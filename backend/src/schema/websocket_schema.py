from datetime import UTC, datetime

from pydantic import BaseModel, ConfigDict, Field

from schema.enums import WSContentType
from schema.shot_schema import Stats


class WebSocketMessage(BaseModel):
    ts: datetime = Field(
        default_factory=lambda: datetime.now(UTC),
        description="Timestamp of the event",
    )
    content_type: WSContentType = Field(WSContentType.SHOT_CREATED)
    content: Stats
    model_config = ConfigDict(populate_by_name=True, extra="forbid")
