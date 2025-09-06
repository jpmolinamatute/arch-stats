from uuid import UUID

import pytest
from asyncpg import Pool

from shared.factories import create_fake_face, create_fake_session
from shared.models import DBException, DBNotFound, FacesModel, SessionsModel, TargetsModel
from shared.schema import FacesFilters, FacesUpdate, TargetsCreate


async def _create_session_and_target(
    db_pool: Pool,
) -> tuple[FacesModel, UUID, float, float, TargetsModel]:
    sessions_db = SessionsModel(db_pool)
    targets_db = TargetsModel(db_pool)
    faces_db = FacesModel(db_pool)
    session_id = await sessions_db.insert_one(create_fake_session())
    target_id = await targets_db.insert_one(
        TargetsCreate(max_x=200.0, max_y=180.0, session_id=session_id, distance=18)
    )
    return faces_db, target_id, 200.0, 180.0, targets_db


@pytest.mark.asyncio
async def test_create_and_get_face(db_pool_initialed: Pool) -> None:
    faces_db, target_id, max_x, max_y, _ = await _create_session_and_target(db_pool_initialed)
    # Safe payload within bounds
    payload = create_fake_face(target_id, max_x, max_y).model_copy(
        update={"human_identifier": "F-1"}
    )
    face_id = await faces_db.insert_one(payload)
    got = await faces_db.get_one_by_id(face_id)
    assert got.human_identifier == "F-1"
    assert isinstance(got.x, float) and isinstance(got.y, float)


@pytest.mark.asyncio
async def test_faces_unique_human_identifier_per_target(db_pool_initialed: Pool) -> None:
    faces_db, target_id, max_x, max_y, _ = await _create_session_and_target(db_pool_initialed)
    hid = "DUP-1"
    base = create_fake_face(target_id, max_x, max_y)
    one = base.model_copy(update={"human_identifier": hid})
    two = base.model_copy(update={"human_identifier": hid})
    await faces_db.insert_one(one)
    with pytest.raises(DBException):
        await faces_db.insert_one(two)


@pytest.mark.asyncio
async def test_faces_array_length_constraints(db_pool_initialed: Pool) -> None:
    faces_db, target_id, _, _, _ = await _create_session_and_target(db_pool_initialed)
    # mismatched lengths -> DBException via CHECK
    bad = create_fake_face(target_id, 200.0, 180.0).model_copy(
        update={"radii": [4.0, 8.0, 12.0], "points": [10, 9]}
    )
    with pytest.raises(DBException):
        await faces_db.insert_one(bad)


@pytest.mark.asyncio
async def test_faces_within_bounds_constraint(db_pool_initialed: Pool) -> None:
    faces_db, target_id, max_x, _, _ = await _create_session_and_target(db_pool_initialed)
    # x + max(radii) >= max_x violates
    # Using strict '>' comparison in validation: equality is allowed; make it exceed bounds
    bad = create_fake_face(target_id, max_x, 180.0).model_copy(
        update={"x": max_x - 3.0, "radii": [4.0], "points": [10]}
    )
    with pytest.raises(DBException):
        await faces_db.insert_one(bad)


@pytest.mark.asyncio
async def test_faces_cascade_delete_on_target(db_pool_initialed: Pool) -> None:
    faces_db, target_id, max_x, max_y, targets_db = await _create_session_and_target(
        db_pool_initialed
    )
    payload = create_fake_face(target_id, max_x, max_y).model_copy(
        update={"human_identifier": "KEEP"}
    )
    face_id = await faces_db.insert_one(payload)
    # Delete parent target
    await targets_db.delete_one(target_id)
    with pytest.raises(DBNotFound):
        await faces_db.get_one_by_id(face_id)


@pytest.mark.asyncio
async def test_get_all_and_filtering_faces(db_pool_initialed: Pool) -> None:
    faces_db, target_id, max_x, max_y, _ = await _create_session_and_target(db_pool_initialed)
    face_a = create_fake_face(target_id, max_x, max_y).model_copy(update={"human_identifier": "AA"})
    face_b = create_fake_face(target_id, max_x, max_y).model_copy(update={"human_identifier": "BB"})
    await faces_db.insert_one(face_a)
    await faces_db.insert_one(face_b)

    all_faces = await faces_db.get_all()
    assert any(f.human_identifier == "AA" for f in all_faces)
    assert any(f.human_identifier == "BB" for f in all_faces)

    only_bb = await faces_db.get_all(FacesFilters(human_identifier="BB"))
    assert len(only_bb) == 1 and only_bb[0].human_identifier == "BB"


@pytest.mark.asyncio
async def test_update_face_and_invalid_update(db_pool_initialed: Pool) -> None:
    faces_db, target_id, max_x, max_y, _ = await _create_session_and_target(db_pool_initialed)
    payload = create_fake_face(target_id, max_x, max_y)
    face_id = await faces_db.insert_one(payload)

    # Valid update
    await faces_db.update_one(face_id, FacesUpdate(x=payload.x + 1.0, human_identifier="NEW"))
    updated = await faces_db.get_one_by_id(face_id)
    assert updated.human_identifier == "NEW"

    # Invalid update (push outside bounds: set huge radii)
    with pytest.raises(DBException):
        await faces_db.update_one(face_id, FacesUpdate(radii=[max_x]))
