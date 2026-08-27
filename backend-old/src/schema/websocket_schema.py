from datetime import UTC, datetime

from pydantic import BaseModel, ConfigDict, Field

from schema.enums import WSContentType
from schema.live_stats_schema import LiveStat


class WebSocketMessage(BaseModel):
    ts: datetime = Field(
        default_factory=lambda: datetime.now(UTC),
        description="Timestamp of the event",
    )
    content_type: WSContentType = Field(WSContentType.SHOT_CREATED)
    content: LiveStat
    model_config = ConfigDict(populate_by_name=True, extra="forbid")
