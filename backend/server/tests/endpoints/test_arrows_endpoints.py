from typing import Any
from uuid import UUID, uuid4

import pytest
from faker import Faker
from httpx import AsyncClient


ARROWS_ENDPOINT = "/api/v0/arrow"


def arrow_payload(**overrides: Any) -> dict[str, object]:
    fake = Faker()
    data = {
        "id": str(uuid4()),
        "length": 29.0,
        "human_identifier": fake.word(),
        "is_programmed": False,
        "label_position": 1.0,
        "weight": 350.0,
        "diameter": 5.7,
        "spine": 400,
    }
    data.update(overrides)
    return data


async def test_arrow_crud_workflow(async_client: AsyncClient, faker: Faker) -> None:
    # --- Create Arrow ---
    arrow_id = str(uuid4())
    human_identifier = faker.word()
    payload = arrow_payload(**{"id": arrow_id, "human_identifier": human_identifier})
    resp = await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    resp_data = resp.json()

    assert resp.status_code == 201
    assert resp_data["code"] == 201
    assert resp_data["errors"] == []

    # --- Get All Arrows ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/")
    assert resp.status_code == 200
    arrows = resp.json()["data"]
    assert isinstance(arrows, list)
    assert any(a["id"] == arrow_id for a in arrows)

    # --- Get Arrow by ID ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code == 200
    arrow = resp.json()["data"]
    assert arrow["id"] == arrow_id
    assert arrow["human_identifier"] == human_identifier

    # --- Update (Patch) Arrow ---
    patch = {
        "length": 30.0,
        "weight": 360.0,
        "label_position": 2.0,
        "is_programmed": True,
        "human_identifier": "A2",
        "diameter": 5.8,
        "spine": 410,
    }
    resp = await async_client.patch(f"{ARROWS_ENDPOINT}/{arrow_id}", json=patch)
    assert resp.status_code == 204
    assert resp.json()["errors"] == []

    # --- Confirm Update ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    updated = resp.json()["data"]
    assert updated["length"] == 30.0
    assert updated["is_programmed"] is True
    assert updated["human_identifier"] == "A2"

    # --- Delete Arrow ---
    resp = await async_client.delete(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code == 204
    assert resp.json()["errors"] == []

    # --- Confirm Gone ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code == 404


async def test_arrow_duplicate_human_identifier(async_client: AsyncClient, faker: Faker) -> None:
    """Should not allow two arrows with the same human_identifier (UNIQUE constraint)."""
    human_identifier = faker.word()
    payload = arrow_payload(**{"id": str(uuid4()), "human_identifier": human_identifier})

    # Insert an arrow the first time
    resp = await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)

    assert resp.status_code == 201
    assert resp.json()["errors"] == []

    # Insert the same arrow a second time
    resp = await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    assert resp.status_code == 400
    assert resp.json()["errors"]


@pytest.mark.parametrize("missing_field", ["id", "length", "human_identifier", "is_programmed"])
async def test_arrow_missing_required_fields(async_client: AsyncClient, missing_field: str) -> None:
    """Should fail when required fields are missing."""
    payload = arrow_payload()
    # Remove the required field
    payload.pop(missing_field, None)
    resp = await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    assert (
        resp.status_code == 422
    ), f"Expected 422 when missing {missing_field}, got {resp.status_code}\n{resp.text}"
    result = resp.json()
    assert "detail" in result
    # Assert that the missing field appears in the error details
    missing_message = (
        f"Expected missing field '{missing_field}' to appear in error detail, got: "
        f"{result['detail']}"
    )
    assert any(missing_field in str(item["loc"]) for item in result["detail"]), missing_message


@pytest.mark.parametrize(
    "field, wrong_value",
    [
        ("id", "not-a-uuid"),
        ("length", "not-a-float"),
        ("human_identifier", 12345),  # expecting str
        ("is_programmed", "not-a-bool"),
        ("label_position", "not-a-float"),
        ("weight", "not-a-float"),
        ("diameter", "not-a-float"),
        ("spine", "not-a-float"),
    ],
)
async def test_arrow_wrong_data_type(
    async_client: AsyncClient, field: str, wrong_value: str
) -> None:
    """Should fail after validation."""
    # Start with a valid payload
    payload = arrow_payload()
    # Replace the field with an invalid value
    payload[field] = wrong_value

    resp = await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    wrong_message = (
        f"Expected 422 for field {field} with value {wrong_value!r}, got "
        f"{resp.status_code}\n{resp.text}"
    )
    assert resp.status_code == 422, wrong_message
    result = resp.json()
    assert "detail" in result
    # Optionally, check that the correct field is mentioned in the error
    assert any(field in str(item["loc"]) for item in result["detail"]), f"{result['detail']}"


async def test_arrow_get_new_id(async_client: AsyncClient) -> None:
    """Should return a valid UUID string from /arrow/new_id."""
    resp = await async_client.get("/api/v0/new_arrow/")
    assert resp.status_code == 200
    new_id = resp.json()
    # Should be a valid UUID
    uuid_obj = UUID(new_id)
    assert str(uuid_obj) == new_id


async def test_patch_arrow_with_no_fields(async_client: AsyncClient) -> None:
    arrow_id = str(uuid4())
    payload = arrow_payload(id=arrow_id)
    await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    resp = await async_client.patch(f"{ARROWS_ENDPOINT}/{arrow_id}", json={})
    assert resp.status_code == 400


async def test_patch_nonexistent_arrow(async_client: AsyncClient) -> None:
    resp = await async_client.patch(f"{ARROWS_ENDPOINT}/{uuid4()}", json={"length": 30.0})
    assert resp.status_code == 404


async def test_delete_nonexistent_arrow(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{ARROWS_ENDPOINT}/{uuid4()}")
    assert resp.status_code == 404


async def test_get_all_arrows_empty(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/")
    assert resp.status_code == 200
    arrows = resp.json()["data"]
    assert arrows == [] or arrows is None


async def test_post_arrow_with_extra_field(async_client: AsyncClient) -> None:
    payload = arrow_payload()
    payload["unexpected_field"] = "should be forbidden"
    resp = await async_client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    # Depending on your config, this could be 422 or 400
    assert resp.status_code == 422
