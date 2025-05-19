import uuid
from pydantic import BaseModel, Field


class ArrowBase(BaseModel):
    human_identifier: str = Field(..., max_length=10)
    is_programmed: bool = False
    weight: float | None = None
    diameter: float | None = None
    spine: float | None = None
    length: float | None = None
    label_position: float | None = None


class ArrowCreate(ArrowBase):
    pass


class ArrowUpdate(BaseModel):
    human_identifier: str | None = Field(None, max_length=10)
    is_programmed: bool | None = None
    weight: float | None = None
    diameter: float | None = None
    spine: float | None = None
    length: float | None = None
    label_position: float | None = None


class ArrowRead(ArrowBase):
    _id: uuid.UUID

    class Config:
        orm_mode = True
