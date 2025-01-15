from typing import TypedDict
from fastapi import APIRouter, HTTPException

router = APIRouter(prefix="/archers", tags=["Archers"])


class Target(TypedDict):
    id: int
    name: str
    prize: str
    start_date: str
    end_date: str


# Simulated database
targets: list[Target] = []


@router.get("/", response_model=list[Target])
async def get_all_tournaments() -> list[Target]:
    return targets


@router.get("/{target_id}", response_model=Target)
async def get_tournament(target_id: int) -> Target:
    for target in targets:
        if target["id"] == target_id:
            return target
    raise HTTPException(status_code=404, detail="Target not found")


@router.post("/", response_model=Target)
async def create_tournament(target: Target) -> Target:
    targets.append(target)
    return target


@router.put("/{target_id}", response_model=Target)
async def update_tournament(target_id: int, updated_tournament: Target) -> Target:
    for idx, target in enumerate(targets):
        if target["id"] == target_id:
            targets[idx] = updated_tournament
            return updated_tournament
    raise HTTPException(status_code=404, detail="Target not found")


@router.delete("/{target_id}", response_model=dict[str, str])
async def delete_tournament(target_id: int) -> dict[str, str]:
    for idx, target in enumerate(targets):
        if target["id"] == target_id:
            del targets[idx]
            return {"message": "Target deleted successfully"}
    raise HTTPException(status_code=404, detail="Target not found")
