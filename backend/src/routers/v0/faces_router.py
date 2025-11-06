from fastapi import APIRouter, HTTPException, status

from core import face_data
from schema import Face, FaceMinimal, FaceType


router = APIRouter(prefix="/faces", tags=["Faces"])


@router.get("", status_code=status.HTTP_200_OK, response_model=list[FaceMinimal])
async def list_faces() -> list[FaceMinimal]:
    faces: list[FaceMinimal] = []
    for f in face_data:
        faces.append(FaceMinimal(face_id=f.face_id, face_name=f.face_name))
    return faces


@router.get("/{face_id}", status_code=status.HTTP_200_OK, response_model=Face)
async def get_face(face_id: FaceType) -> Face:
    for f in face_data:
        if f.face_id == face_id:
            return f
    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND,
        detail=f"Face with id '{face_id}' not found",
    )
