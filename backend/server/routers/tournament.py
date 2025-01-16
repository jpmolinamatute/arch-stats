from typing import TypedDict

from fastapi import APIRouter, HTTPException


router = APIRouter(prefix="/tournaments", tags=["Tournaments"])


class Tournament(TypedDict):
    id: int
    name: str
    prize: str
    start_date: str
    end_date: str


# Simulated database
tournaments: list[Tournament] = []


@router.get("/", response_model=list[Tournament])
async def get_all_tournaments() -> list[Tournament]:
    return tournaments


@router.get("/{tournament_id}", response_model=Tournament)
async def get_tournament(tournament_id: int) -> Tournament:
    for tournament in tournaments:
        if tournament["id"] == tournament_id:
            return tournament
    raise HTTPException(status_code=404, detail="Tournament not found")


@router.post("/", response_model=Tournament)
async def create_tournament(tournament: Tournament) -> Tournament:
    tournaments.append(tournament)
    return tournament


@router.put("/{tournament_id}", response_model=Tournament)
async def update_tournament(tournament_id: int, updated_tournament: Tournament) -> Tournament:
    for idx, tournament in enumerate(tournaments):
        if tournament["id"] == tournament_id:
            tournaments[idx] = updated_tournament
            return updated_tournament
    raise HTTPException(status_code=404, detail="Tournament not found")


@router.delete("/{tournament_id}", response_model=dict[str, str])
async def delete_tournament(tournament_id: int) -> dict[str, str]:
    for idx, tournament in enumerate(tournaments):
        if tournament["id"] == tournament_id:
            del tournaments[idx]
            return {"message": "Tournament deleted successfully"}
    raise HTTPException(status_code=404, detail="Tournament not found")
