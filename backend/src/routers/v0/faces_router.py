from fastapi import APIRouter, HTTPException, status

from core import face_data
from schema import Face, FaceMinimal, FaceType


router = APIRouter(prefix="/faces", tags=["Faces"])


@router.get("", status_code=status.HTTP_200_OK, response_model=list[FaceMinimal])
async def list_faces() -> list[FaceMinimal]:
    faces: list[FaceMinimal] = []
    for f in face_data:
        faces.append(FaceMinimal(face_type=f.face_type, face_name=f.face_name))
    return faces


@router.get("/{face_type}", status_code=status.HTTP_200_OK, response_model=Face)
async def get_face(face_type: FaceType) -> Face:
    for f in face_data:
        if f.face_type == face_type:
            return f
    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND,
        detail=f"Face with type '{face_type}' not found",
    )
