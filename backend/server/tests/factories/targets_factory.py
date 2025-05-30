from typing import Any
from uuid import UUID

from httpx import AsyncClient

from server.schema import TargetsCreate, TargetsRead

TARGETS_ENDPOINT = "/api/v0/target"


def create_fake_target(session_id: UUID, **overrides: Any) -> TargetsCreate:
    """
    Generate a valid payload for TargetsCreate.
    Adjust the fields to match your schema exactly!
    """

    data = TargetsCreate(
        max_x_coordinate=122.0,
        max_y_coordinate=122.0,
        radius=[3.0, 6.0, 9.0, 12.0, 15.0],
        points=[10, 8, 6, 4, 2],
        height=140.0,
        session_id=session_id,
        human_identifier="T1",
    )
    return data.model_copy(update=overrides)


async def create_many_targets(
    async_client: AsyncClient, session_id: UUID, count: int = 5
) -> list[TargetsRead]:
    targets = []
    for i in range(count):
        payload = create_fake_target(
            session_id=session_id,
            max_x_coordinate=120.0 + i,
            max_y_coordinate=121.0 + i,
            height=140.0 + i,
            human_identifier=f"target_{i}",
        )
        payload_dict = payload.model_dump(mode="json", by_alias=True)
        resp = await async_client.post(TARGETS_ENDPOINT, json=payload_dict)
        assert resp.status_code == 201
        payload_dict["id"] = resp.json()["data"]
        targets.append(TargetsRead(**payload_dict))
    return targets
