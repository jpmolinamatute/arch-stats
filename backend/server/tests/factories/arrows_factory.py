import random
from typing import Any
from uuid import uuid4, UUID

from httpx import AsyncClient

from server.schema import ArrowsCreate

ARROWS_ENDPOINT = "/api/v0/arrow"


def create_fake_arrow(**overrides: Any) -> ArrowsCreate:
    ran_int = random.randint(1, 100)
    data = ArrowsCreate(
        id=uuid4(),
        length=29.0,
        human_identifier=f"arrow-{ran_int}",
        is_programmed=False,
        label_position=1.0,
        weight=350.0,
        diameter=5.7,
        spine=400,
    )
    return data.model_copy(update=overrides)


async def create_many_arrows(async_client: AsyncClient, count: int = 5) -> list[UUID]:
    arrows = []
    for i in range(count):
        arrow = create_fake_arrow(
            is_programmed=bool(i % 2),
            spine=400 + 10 * i,
            weight=350.0 + 5 * i,
            human_identifier=f"arrow_{i}",
            length=29.0 + i,
            label_position=1.0 + 0.5 * i,
            diameter=5.7 + 0.1 * i,
        )
        payload = arrow.model_dump(mode="json", by_alias=True)
        resp = await async_client.post(ARROWS_ENDPOINT, json=payload)

        assert resp.status_code == 201
        _id = payload["id"]
        arrows.append(_id)
    return arrows
