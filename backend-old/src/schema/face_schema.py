from pydantic import BaseModel, ConfigDict, Field

from schema.enums import FaceType


class Spot(BaseModel):
    """
    We use offset from center to allocate multiple spots on the same face in relation to
    the canvas center. At the same time, x_offset and y_offset are the center of the spot.
    """

    x_offset: float = Field(..., description="in mm, offset from center")
    y_offset: float = Field(..., description="in mm, offset from center")
    diameter: float = Field(..., description="in mm")
    model_config = ConfigDict(title="Spot Schema", extra="forbid")


class Ring(BaseModel):
    data_score: int = Field(..., description="e.g., 10, 9, 8, ...")
    fill: str = Field(..., description="Color of the scoring zone")
    r: float = Field(..., description="Radius of the scoring zone")
    stroke: str = Field(..., description="Color of the outer line")
    stroke_width: float = Field(..., description="Thickness of the outer line")

    model_config = ConfigDict(title="Ring Schema", extra="forbid")


class FaceMinimal(BaseModel):
    face_type: FaceType = Field(..., description="Unique identifier for the face")
    face_name: str = Field(..., description="Name of the face")
    model_config = ConfigDict(title="Face Minimal Schema", extra="forbid")


class Face(FaceMinimal):
    spots: list[Spot] = Field(..., description="List of spots on the face")
    rings: list[Ring] = Field(..., description="List of rings on the face")
    viewBox: float = Field(..., description="SVG viewBox size")
    render_cross: bool = Field(
        default=False, description="Whether to render the cross in the center"
    )
    model_config = ConfigDict(title="Face Schema", extra="forbid")
