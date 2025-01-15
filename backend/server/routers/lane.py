from typing import TypedDict
from fastapi import APIRouter, HTTPException

router = APIRouter(prefix="/lanes", tags=["Lanes"])


class Lane(TypedDict):
    id: int
    name: str
    prize: str
    start_date: str
    end_date: str


# Simulated database
lanes: list[Lane] = []


@router.get("/", response_model=list[Lane])
async def get_all_tournaments() -> list[Lane]:
    return lanes


@router.get("/{lane_id}", response_model=Lane)
async def get_tournament(lane_id: int) -> Lane:
    for lane in lanes:
        if lane["id"] == lane_id:
            return lane
    raise HTTPException(status_code=404, detail="Lane not found")


@router.post("/", response_model=Lane)
async def create_tournament(lane: Lane) -> Lane:
    lanes.append(lane)
    return lane


@router.put("/{lane_id}", response_model=Lane)
async def update_tournament(lane_id: int, updated_tournament: Lane) -> Lane:
    for idx, lane in enumerate(lanes):
        if lane["id"] == lane_id:
            lanes[idx] = updated_tournament
            return updated_tournament
    raise HTTPException(status_code=404, detail="Lane not found")


@router.delete("/{lane_id}", response_model=dict[str, str])
async def delete_tournament(lane_id: int) -> dict[str, str]:
    for idx, lane in enumerate(lanes):
        if lane["id"] == lane_id:
            del lanes[idx]
            return {"message": "Lane deleted successfully"}
    raise HTTPException(status_code=404, detail="Lane not found")
