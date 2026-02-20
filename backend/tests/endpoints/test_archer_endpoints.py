import datetime
from http import HTTPStatus
from uuid import uuid4

import pytest
from httpx import AsyncClient

from schema import ArcherCreate, BowStyleType, GenderType


@pytest.fixture
def sample_archer_data() -> ArcherCreate:
    return ArcherCreate(
        first_name="Test",
        last_name="Archer",
        email="test.archer@example.com",
        date_of_birth=datetime.date(1990, 1, 1),
        gender=GenderType.UNSPECIFIED,
        bowstyle=BowStyleType.RECURVE,
        draw_weight=40.0,
        google_subject="test_subject_123",
        google_picture_url="",
    )


@pytest.mark.asyncio
async def test_create_archer_success(client: AsyncClient, sample_archer_data: ArcherCreate) -> None:
    response = await client.post("/api/v0/archer/", json=sample_archer_data.model_dump(mode="json"))
    assert response.status_code == HTTPStatus.CREATED
    data = response.json()
    assert "archer_id" in data


@pytest.mark.asyncio
async def test_list_archers(client: AsyncClient, sample_archer_data: ArcherCreate) -> None:
    # Create an archer first
    await client.post("/api/v0/archer/", json=sample_archer_data.model_dump(mode="json"))

    response = await client.get("/api/v0/archer/")
    assert response.status_code == HTTPStatus.OK
    data = response.json()
    assert len(data) >= 1
    assert any(a["first_name"] == "Test" for a in data)


@pytest.mark.asyncio
async def test_get_archer_success(client: AsyncClient, sample_archer_data: ArcherCreate) -> None:
    create_response = await client.post(
        "/api/v0/archer/", json=sample_archer_data.model_dump(mode="json")
    )
    archer_id = create_response.json()["archer_id"]

    response = await client.get(f"/api/v0/archer/{archer_id}")
    assert response.status_code == HTTPStatus.OK
    data = response.json()
    assert data["archer_id"] == str(archer_id)
    assert data["first_name"] == "Test"


@pytest.mark.asyncio
async def test_get_archer_not_found(client: AsyncClient) -> None:
    random_id = uuid4()
    response = await client.get(f"/api/v0/archer/{random_id}")
    assert response.status_code == HTTPStatus.NOT_FOUND


@pytest.mark.asyncio
async def test_update_archer_success(client: AsyncClient, sample_archer_data: ArcherCreate) -> None:
    create_response = await client.post(
        "/api/v0/archer/", json=sample_archer_data.model_dump(mode="json")
    )
    archer_id = create_response.json()["archer_id"]

    update_data = {
        "where": {"archer_id": archer_id},
        "data": {"first_name": "UpdatedName"},
    }
    response = await client.patch("/api/v0/archer/", json=update_data)
    assert response.status_code == HTTPStatus.OK

    get_response = await client.get(f"/api/v0/archer/{archer_id}")
    assert get_response.json()["first_name"] == "UpdatedName"


@pytest.mark.asyncio
async def test_delete_archer_success(client: AsyncClient, sample_archer_data: ArcherCreate) -> None:
    create_response = await client.post(
        "/api/v0/archer/", json=sample_archer_data.model_dump(mode="json")
    )
    archer_id = create_response.json()["archer_id"]

    delete_response = await client.delete(f"/api/v0/archer/{archer_id}")
    assert delete_response.status_code == HTTPStatus.NO_CONTENT

    get_response = await client.get(f"/api/v0/archer/{archer_id}")
    assert get_response.status_code == HTTPStatus.NOT_FOUND
