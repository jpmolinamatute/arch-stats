from typing import TypedDict
from uuid import UUID

from fastapi import APIRouter, HTTPException, Query


router = APIRouter(prefix="/shooting", tags=["Shooting"])


class Shooting(TypedDict):
    id: int
    name: str
    prize: str
    start_date: str
    end_date: str


# Simulated database
shootings: list[Shooting] = []


@router.get("/", response_model=list[Shooting])
async def get_all_shootings(
    archer_id: UUID = Query(None, description="Archer's UUID")
) -> list[Shooting]:
    return shootings


@router.get("/{shooting_id}", response_model=Shooting)
async def get_shooting(shooting_id: int) -> Shooting:
    for shooting in shootings:
        if shooting["id"] == shooting_id:
            return shooting
    raise HTTPException(status_code=404, detail="Shooting not found")


@router.post("/", response_model=Shooting)
async def create_shooting(shooting: Shooting) -> Shooting:
    shootings.append(shooting)
    return shooting


@router.put("/{shooting_id}", response_model=Shooting)
async def update_shooting(shooting_id: int, updated_shooting: Shooting) -> Shooting:
    for idx, shooting in enumerate(shootings):
        if shooting["id"] == shooting_id:
            shootings[idx] = updated_shooting
            return updated_shooting
    raise HTTPException(status_code=404, detail="Shooting not found")


@router.delete("/{shooting_id}", response_model=dict[str, str])
async def delete_shooting(shooting_id: int) -> dict[str, str]:
    for idx, shooting in enumerate(shootings):
        if shooting["id"] == shooting_id:
            del shootings[idx]
            return {"message": "Shooting deleted successfully"}
    raise HTTPException(status_code=404, detail="Shooting not found")
