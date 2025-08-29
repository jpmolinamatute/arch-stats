import json
from uuid import uuid4

import pytest
from asyncpg import Pool

from shared.factories import create_fake_target, create_many_sessions
from shared.models import DBException, DBNotFound, TargetsDB


@pytest.mark.asyncio
async def test_create_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id, face_count=1)
    payload.faces[0].human_identifier = "T01"
    target_id = await db.insert_one(payload)
    assert target_id is not None
    target = await db.get_one_by_id(target_id)
    assert target is not None
    assert target.faces[0].human_identifier == "T01"


@pytest.mark.asyncio
async def test_get_all_targets(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload1 = create_fake_target(session[0].session_id, face_count=1)
    payload1.faces[0].human_identifier = "T02"
    payload2 = create_fake_target(session[0].session_id, face_count=1)
    payload2.faces[0].human_identifier = "T03"
    await db.insert_one(payload1)
    await db.insert_one(payload2)
    targets = await db.get_all()
    identifiers = [face.human_identifier for t in targets for face in t.faces]
    assert "T02" in identifiers and "T03" in identifiers


@pytest.mark.asyncio
async def test_get_specific_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id, face_count=1, human_identifier="T04")
    target_id = await db.insert_one(payload)
    fetched = await db.get_one({"id": target_id})
    assert fetched is not None
    assert target_id == fetched.get_id()


@pytest.mark.asyncio
async def test_delete_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id, face_count=1)
    target_id = await db.insert_one(payload)
    await db.delete_one(target_id)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(target_id)


@pytest.mark.asyncio
async def test_get_one_by_id_target(db_pool_initialed: Pool) -> None:
    session = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)
    payload = create_fake_target(session[0].session_id, face_count=1)
    payload.faces[0].human_identifier = "T99"
    target_id = await db.insert_one(payload)
    target = await db.get_one_by_id(target_id)
    assert target.faces[0].human_identifier == "T99"


@pytest.mark.asyncio
async def test_get_nonexistent_target_raises(db_pool_initialed: Pool) -> None:
    db = TargetsDB(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await db.get_one_by_id(uuid4())


@pytest.mark.asyncio
async def test_delete_nonexistent_target_raises(db_pool_initialed: Pool) -> None:
    db = TargetsDB(db_pool_initialed)
    with pytest.raises(DBNotFound):
        await db.delete_one(uuid4())


@pytest.mark.asyncio
async def test_get_by_session_id(db_pool_initialed: Pool) -> None:
    sessions = await create_many_sessions(db_pool_initialed, 2)
    db = TargetsDB(db_pool_initialed)
    t1 = create_fake_target(sessions[0].session_id, face_count=1, human_identifier="S1-T1")
    t2 = create_fake_target(sessions[0].session_id, face_count=1, human_identifier="S1-T2")
    t3 = create_fake_target(sessions[1].session_id, face_count=1, human_identifier="S2-T1")
    await db.insert_one(t1)
    await db.insert_one(t2)
    await db.insert_one(t3)

    s1_targets = await db.get_by_session_id(sessions[0].session_id)
    assert len(s1_targets) == 2
    assert all(t.session_id == sessions[0].session_id for t in s1_targets)


@pytest.mark.asyncio
async def test_jsonb_validation_rejects_invalid_face(db_pool_initialed: Pool) -> None:
    # x must be > 0 and radii must be > 0 per DB validation function
    sessions = await create_many_sessions(db_pool_initialed, 1)
    db = TargetsDB(db_pool_initialed)

    # Bypass Pydantic for 'faces' by inserting with raw SQL + JSONB cast.
    faces_payload = [
        {
            "x": 0.0,  # invalid per CHECK
            "y": 10.0,
            "radii": [-10.0],  # invalid per CHECK
            "points": [10],
            "human_identifier": "BAD",
        }
    ]
    faces_json = json.dumps(faces_payload)

    insert_sql = (
        "INSERT INTO targets (max_x, max_y, session_id, faces) VALUES ($1, $2, $3, $4::jsonb);"
    )

    with pytest.raises(DBException):
        await db.execute(
            insert_sql,
            [100.0, 100.0, sessions[0].session_id, faces_json],
        )
