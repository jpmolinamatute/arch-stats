from math import dist
from typing import TypedDict
from uuid import UUID
from datetime import datetime


class SensorData(TypedDict, total=False):
    id: UUID | None
    target_track_id: UUID
    arrow_id: UUID
    arrow_engage_time: datetime
    draw_length: float
    arrow_disengage_time: datetime
    arrow_landing_time: datetime | None
    x_coordinate: float | None
    y_coordinate: float | None
    distance: float


SensorDataTuple = tuple[
    UUID, UUID, datetime, float, datetime, datetime | None, float | None, float | None, float
]
