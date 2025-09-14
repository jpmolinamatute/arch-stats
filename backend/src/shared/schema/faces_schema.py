from typing import Annotated
from uuid import UUID

from annotated_types import Len
from pydantic import BaseModel, ConfigDict, Field


class FaceCalibration(BaseModel):
    x: float = Field(..., description="X coordinate of face center")
    y: float = Field(..., description="Y coordinate of face center")
    radii: Annotated[list[float], Len(min_length=1)] = Field(
        default_factory=list, description="List of radii for the face rings"
    )


class Face(BaseModel):
    x: float = Field(..., description="X coordinate of face center")
    y: float = Field(..., description="Y coordinate of face center")
    human_identifier: str = Field(..., description="Unique identifier within target")
    radii: Annotated[list[float], Len(min_length=1)] = Field(
        default_factory=list, description="List of radii for the face rings"
    )
    points: Annotated[list[int], Len(min_length=1)] = Field(
        default_factory=list, description="Points assigned to the face rings"
    )
    model_config = ConfigDict(extra="forbid")


class FacesCreate(BaseModel):
    target_id: UUID = Field(..., description="Parent target id")
    x: float = Field(..., description="X coordinate of face center")
    y: float = Field(..., description="Y coordinate of face center")
    human_identifier: str = Field(..., description="Unique identifier within target")
    radii: Annotated[list[float], Len(min_length=1)] = Field(
        default_factory=list, description="List of radii for the face rings"
    )
    points: Annotated[list[int], Len(min_length=1)] = Field(
        default_factory=list, description="Points assigned to the face rings"
    )
    model_config = ConfigDict(extra="forbid")


class FacesUpdate(BaseModel):
    x: float | None = Field(default=None, description="X coordinate of face center")
    y: float | None = Field(default=None, description="Y coordinate of face center")
    human_identifier: str | None = Field(
        default=None, description="Unique identifier within target"
    )
    radii: list[float] | None = Field(default=None, description="Rings radii")
    points: list[int] | None = Field(default=None, description="Rings points")
    model_config = ConfigDict(extra="forbid")


class FacesFilters(BaseModel):
    face_id: UUID | None = Field(default=None, alias="id", description="ID of the face")
    target_id: UUID | None = Field(default=None, description="Parent target id")
    human_identifier: str | None = Field(default=None, description="Identifier filter")
    model_config = ConfigDict(extra="forbid")


class FacesRead(FacesCreate):
    face_id: UUID = Field(..., alias="id")
    model_config = ConfigDict(extra="forbid")

    def get_id(self) -> UUID:
        return self.face_id
