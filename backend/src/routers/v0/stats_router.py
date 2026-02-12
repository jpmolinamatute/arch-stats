from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, WebSocket, WebSocketDisconnect, status

from models import DBNotFound, ShotModel
from routers.deps.models import get_shot_model, get_shot_model_ws
from schema import Stats, WebSocketMessage, WSContentType

router = APIRouter(prefix="/stats", tags=["Stats"])


@router.get("/{slot_id:uuid}", response_model=Stats, status_code=status.HTTP_200_OK)
async def get_stats(
    slot_id: UUID,
    shot_model: Annotated[ShotModel, Depends(get_shot_model)],
) -> Stats:
    """
    Get live statistics and shots for a slot.
    """
    try:
        # 1. Get live stats
        live_stat = await shot_model.get_live_stat(slot_id)

        # 2. Get all shots for the slot to build the list of ShotScore
        # We use get_scores to retrieve only shot_id and score
        shots_scores = await shot_model.get_scores(slot_id)

        # 3. Return Stats
        return Stats(shots=shots_scores, stats=live_stat)

    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(e)) from e


@router.websocket("/ws/{slot_id:uuid}")
async def websocket_stats(
    websocket: WebSocket,
    slot_id: UUID,
    shot_model: ShotModel = Depends(get_shot_model_ws),
) -> None:
    """WebSocket endpoint streaming shot notifications for a slot.

    Relays DB NOTIFY payloads from channel "shot_insert_{slot_id}" to the client
    as JSON when possible; falls back to raw text otherwise.
    """

    await websocket.accept()
    try:
        async for payload in shot_model.listen_for_shots(slot_id):
            message = WebSocketMessage(content=payload, content_type=WSContentType.SHOT_CREATED)
            await websocket.send_json(message.model_dump(mode="json"))
    except WebSocketDisconnect:
        # Client disconnected; the generator will be cancelled and listener removed
        pass
