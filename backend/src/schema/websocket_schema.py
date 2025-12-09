from __future__ import annotations

from datetime import datetime, timezone
from typing import Annotated, Literal, Union

from pydantic import BaseModel, ConfigDict, Field

from schema.shot_schema import Stats


class WebSocketMessageBase(BaseModel):
    ts: datetime = Field(
        default_factory=lambda: datetime.now(timezone.utc),
        description="Timestamp of the event",
    )
    model_config = ConfigDict(populate_by_name=True)


class ShotCreatedMessage(WebSocketMessageBase):
    data_type: Literal["shot.created"] = "shot.created"
    data_content: Stats


WebSocketMessage = Annotated[Union[ShotCreatedMessage], Field(discriminator="data_type")]
