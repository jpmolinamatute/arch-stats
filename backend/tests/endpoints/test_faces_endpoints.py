from http import HTTPStatus

import pytest
from httpx import AsyncClient

from core import face_data


@pytest.mark.asyncio
async def test_list_faces(client: AsyncClient) -> None:
    response = await client.get("/api/v0/faces")
    assert response.status_code == HTTPStatus.OK
    data = response.json()

    assert len(data) == len(face_data)

    for item in data:
        assert "face_type" in item
        assert "face_name" in item


@pytest.mark.asyncio
async def test_get_face_success(client: AsyncClient) -> None:
    first_face = face_data[0]
    face_type = first_face.face_type.value

    response = await client.get(f"/api/v0/faces/{face_type}")
    assert response.status_code == HTTPStatus.OK
    data = response.json()

    assert data["face_type"] == face_type
    assert data["rings"] == first_face.rings


@pytest.mark.asyncio
async def test_get_face_not_found(client: AsyncClient) -> None:
    response = await client.get("/api/v0/faces/non_existent_face")
    assert response.status_code == HTTPStatus.UNPROCESSABLE_ENTITY
