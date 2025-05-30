from uuid import UUID, uuid4
from asyncpg import Pool
import pytest
from faker import Faker
from httpx import AsyncClient
from server.tests.factories import create_fake_arrow, create_many_arrows

ARROWS_ENDPOINT = "/api/v0/arrow"


@pytest.mark.asyncio
async def test_arrow_crud_workflow(async_client: AsyncClient, faker: Faker) -> None:
    # --- Create Arrow ---
    arrow_uuid = uuid4()
    arrow_id = str(arrow_uuid)
    human_identifier = faker.word()
    arrow = create_fake_arrow(arrow_id=arrow_uuid, human_identifier=human_identifier)
    payload_dict = arrow.model_dump(mode="json", by_alias=True)
    resp = await async_client.post(f"{ARROWS_ENDPOINT}", json=payload_dict)
    resp_data = resp.json()

    assert resp.status_code == 201
    assert resp_data["code"] == 201
    assert resp_data["data"] == arrow_id
    assert resp_data["errors"] == []

    # # --- Get All Arrows ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}")
    assert resp.status_code == 200
    arrows = resp.json()["data"]
    assert isinstance(arrows, list)
    assert any(a["id"] == arrow_id for a in arrows)

    # --- Get Arrow by ID ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code == 200
    arrows = resp.json()["data"]
    assert arrows["id"] == arrow_id
    assert arrows["human_identifier"] == human_identifier

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
    assert resp.status_code == 202
    assert resp.json()["errors"] == []

    # --- Confirm Update ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    resp_json = resp.json()
    updated = resp_json["data"]
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


@pytest.mark.asyncio
async def test_arrow_duplicate_human_identifier(async_client: AsyncClient, faker: Faker) -> None:
    """Should not allow two arrows with the same human_identifier (UNIQUE constraint)."""
    human_identifier = faker.word()
    arrow = create_fake_arrow(
        id=uuid4(),
        human_identifier=human_identifier,
    )
    payload = arrow.model_dump(mode="json", by_alias=True)
    # Insert an arrow the first time
    resp = await async_client.post(f"{ARROWS_ENDPOINT}", json=payload)

    assert resp.status_code == 201
    assert resp.json()["errors"] == []

    # Insert the same arrow a second time
    resp = await async_client.post(f"{ARROWS_ENDPOINT}", json=payload)
    assert resp.status_code == 400
    assert resp.json()["errors"]


@pytest.mark.asyncio
@pytest.mark.parametrize("missing_field", ["id", "length", "human_identifier", "is_programmed"])
async def test_arrow_missing_required_fields(async_client: AsyncClient, missing_field: str) -> None:
    """Should fail when required fields are missing."""
    arrow = create_fake_arrow()
    # Remove the required field
    payload_dict = arrow.model_dump(mode="json", by_alias=True)
    payload_dict.pop(missing_field, None)
    resp = await async_client.post(f"{ARROWS_ENDPOINT}", json=payload_dict)
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


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "field, wrong_value",
    [
        ("id", "not-a-uuid"),
        ("length", "not-a-float"),
        ("human_identifier", 12345),
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
    arrow = create_fake_arrow()
    payload_dict = arrow.model_dump(mode="json", by_alias=True)
    # Replace the field with an invalid value
    payload_dict[field] = wrong_value

    resp = await async_client.post(f"{ARROWS_ENDPOINT}", json=payload_dict)
    wrong_message = (
        f"Expected 422 for field {field} with value {wrong_value!r}, got "
        f"{resp.status_code}\n{resp.text}"
    )
    assert resp.status_code == 422, wrong_message
    result = resp.json()
    assert "detail" in result
    assert any(field in str(item["loc"]) for item in result["detail"]), f"{result['detail']}"


@pytest.mark.asyncio
async def test_arrow_get_new_id(async_client: AsyncClient) -> None:
    """Should return a valid UUID string from /arrow/new_arrow_uuid."""
    resp = await async_client.get("/api/v0/arrow/new_arrow_uuid")
    assert resp.status_code == 200
    resp_json = resp.json()
    # Should be a valid UUID
    new_id = resp_json["data"]
    uuid_obj = UUID(new_id)
    assert str(uuid_obj) == new_id


@pytest.mark.asyncio
async def test_patch_arrow_with_no_fields(async_client: AsyncClient) -> None:
    arrow_id = str(uuid4())
    arrow = create_fake_arrow(id=arrow_id)
    payload_dict = arrow.model_dump(mode="json", by_alias=True)
    await async_client.post(f"{ARROWS_ENDPOINT}", json=payload_dict)
    resp = await async_client.patch(f"{ARROWS_ENDPOINT}/{arrow_id}", json={})
    assert resp.status_code == 400


@pytest.mark.asyncio
async def test_patch_nonexistent_arrow(async_client: AsyncClient) -> None:
    resp = await async_client.patch(f"{ARROWS_ENDPOINT}/{uuid4()}", json={"length": 30.0})
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_delete_nonexistent_arrow(async_client: AsyncClient) -> None:
    resp = await async_client.delete(f"{ARROWS_ENDPOINT}/{uuid4()}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_get_all_arrows_empty(async_client: AsyncClient) -> None:
    resp = await async_client.get(f"{ARROWS_ENDPOINT}")
    assert resp.status_code == 200
    arrows = resp.json()["data"]
    assert arrows == [] or arrows is None


@pytest.mark.asyncio
async def test_post_arrow_with_extra_field(async_client: AsyncClient) -> None:
    arrow = create_fake_arrow()
    payload_dict = arrow.model_dump(mode="json", by_alias=True)
    payload_dict["unexpected_field"] = "should be forbidden"
    resp = await async_client.post(f"{ARROWS_ENDPOINT}", json=payload_dict)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_arrows_filtering(async_client: AsyncClient, db_pool: Pool) -> None:
    """Test filtering arrows by various fields."""
    # Create multiple arrows
    await create_many_arrows(db_pool, 6)

    # --- Test filter by is_programmed ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}?is_programmed=true")
    resp_json = resp.json()
    data = resp_json["data"]

    assert resp.status_code == 200
    # Only arrows with is_programmed=True
    assert all(a["is_programmed"] is True for a in data)

    resp = await async_client.get(f"{ARROWS_ENDPOINT}?is_programmed=false")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(a["is_programmed"] is False for a in data)

    # --- Test filter by specific spine value ---
    spine_value = 410
    resp = await async_client.get(f"{ARROWS_ENDPOINT}?spine={spine_value}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(a["spine"] == spine_value for a in data)

    # --- Test filter by human_identifier ---
    hid = "arrow_3"
    resp = await async_client.get(f"{ARROWS_ENDPOINT}?human_identifier={hid}")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert len(data) == 1
    assert data[0]["human_identifier"] == hid

    # --- Test filter by multiple fields ---
    resp = await async_client.get(f"{ARROWS_ENDPOINT}?spine=420&is_programmed=true")
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert all(a["spine"] == 420 and a["is_programmed"] is True for a in data)
