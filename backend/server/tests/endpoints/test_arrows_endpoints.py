import uuid
from typing import Any

from fastapi.testclient import TestClient
from faker import Faker

ARROWS_ENDPOINT = "/api/v0/arrow"


def arrow_payload(arrow_id: str | None = None, **overrides: Any) -> dict[str, object]:
    data = {
        "id": str(arrow_id) if arrow_id else str(uuid.uuid4()),
        "length": 29.0,
        "human_identifier": "A1",
        "is_programmed": False,
        "label_position": 1.0,
        "weight": 350.0,
        "diameter": 5.7,
        "spine": 400,
    }
    data.update(overrides)
    return data


def test_arrow_crud_workflow(client: TestClient) -> None:
    # --- Create Arrow ---
    arrow_id = str(uuid.uuid4())
    payload = arrow_payload(arrow_id)
    resp = client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    assert resp.status_code in (200, 201, 204)
    assert resp.json()["code"] in (200, 201, 204)
    assert resp.json()["errors"] == []

    # --- Get All Arrows ---
    resp = client.get(f"{ARROWS_ENDPOINT}/")
    assert resp.status_code == 200
    arrows = resp.json()["data"]
    assert isinstance(arrows, list)
    assert any(a["id"] == arrow_id for a in arrows)

    # --- Get Arrow by ID ---
    resp = client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code == 200
    arrow = resp.json()["data"]
    assert arrow["id"] == arrow_id
    assert arrow["human_identifier"] == "A1"

    # --- Update (Patch) Arrow ---
    patch = {
        "length": 30.0,
        "weight": 360.0,
        "label_position": 2.0,
        "is_programmed": True,
        "human_identifier": "A1",
        "diameter": 5.8,
        "spine": 410,
    }
    resp = client.patch(f"{ARROWS_ENDPOINT}/{arrow_id}", json=patch)
    assert resp.status_code in (202, 200, 204)
    assert resp.json()["errors"] == []

    # --- Confirm Update ---
    resp = client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    updated = resp.json()["data"]
    assert updated["length"] == 30.0
    assert updated["is_programmed"] is True

    # --- Delete Arrow ---
    resp = client.delete(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code in (204, 200)
    assert resp.json()["errors"] == []

    # --- Confirm Gone ---
    resp = client.get(f"{ARROWS_ENDPOINT}/{arrow_id}")
    assert resp.status_code == 404


def test_arrow_duplicate_human_identifier(client: TestClient, faker: Faker) -> None:
    """Should not allow two arrows with the same human_identifier (UNIQUE constraint)."""
    human_identifier = faker.word()
    print(f"This is the name = {human_identifier}")
    payload = arrow_payload(str(uuid.uuid4()), human_identifier=human_identifier)

    # Insert an arrow the first time
    resp = client.post(f"{ARROWS_ENDPOINT}/", json=payload)

    assert resp.status_code == 201
    print(f"This is the response {resp.json()}")
    assert resp.json()["errors"] == []

    # Insert the same arrow a second time
    resp = client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    assert resp.status_code == 400
    assert resp.json()["errors"]


def test_arrow_missing_required_fields(client: TestClient) -> None:
    """Should fail when required fields are missing."""
    payload = {
        "id": str(uuid.uuid4()),
        # Missing length and human_identifier, is_programmed!
    }
    resp = client.post(f"{ARROWS_ENDPOINT}/", json=payload)
    assert resp.status_code == 422
    result = resp.json()
    assert len(result["detail"]) == 3
    for detail in result["detail"]:
        for loc in detail["loc"]:
            if loc != "body":
                assert loc in ("is_programmed", "length", "human_identifier")


def test_arrow_get_new_id(client: TestClient) -> None:
    """Should return a valid UUID string from /arrow/new_id."""
    resp = client.get("/api/v0/new_arrow/")
    assert resp.status_code == 200
    new_id = resp.json()
    # Should be a valid UUID
    uuid_obj = uuid.UUID(new_id)
    assert str(uuid_obj) == new_id
