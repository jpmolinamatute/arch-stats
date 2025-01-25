from math import dist
from typing import TypedDict
from uuid import UUID
from datetime import datetime


class SensorData(TypedDict, total=False):
    id: UUID | None
    target_track_id: UUID
    arrow_id: UUID
    arrow_engage_time: datetime
    arrow_disengage_time: datetime
    arrow_landing_time: datetime | None
    draw_length: float
    x_coordinate: float | None
    y_coordinate: float | None
    distance: float


SensorDataTuple = tuple[
    UUID, UUID, datetime, datetime, datetime | None, float, float | None, float | None, float
]
