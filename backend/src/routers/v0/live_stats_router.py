from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, WebSocket, WebSocketDisconnect, status

from models import DBNotFound, LiveStatsModel
from routers.deps.models import get_live_stats_model, get_live_stats_model_ws
from schema import WebSocketMessage, WSContentType
from schema.live_stats_schema import LiveStat

router = APIRouter(prefix="/stats", tags=["Stats"])


@router.get("/{slot_id:uuid}", response_model=LiveStat, status_code=status.HTTP_200_OK)
async def get_stats(
    slot_id: UUID,
    live_stats_model: Annotated[LiveStatsModel, Depends(get_live_stats_model)],
) -> LiveStat:
    """
    Get live statistics and shots for a slot.
    """
    try:
        live_stat = await live_stats_model.get_live_stat(slot_id)
        shots_scores = await live_stats_model.get_scores(slot_id)
        return LiveStat(shots=shots_scores, stats=live_stat)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(e)) from e


@router.websocket("/ws/{slot_id:uuid}")
async def websocket_stats(
    websocket: WebSocket,
    slot_id: UUID,
    live_stats_model: Annotated[LiveStatsModel, Depends(get_live_stats_model_ws)],
) -> None:
    """WebSocket endpoint streaming shot notifications for a slot.

    Relays DB NOTIFY payloads from channel "shot_insert_{slot_id}" to the client
    as JSON when possible; falls back to raw text otherwise.
    """

    await websocket.accept()
    try:
        async for payload in live_stats_model.listen_for_shots(slot_id):
            message = WebSocketMessage(content=payload, content_type=WSContentType.SHOT_CREATED)
            await websocket.send_json(message.model_dump(mode="json"))
    except WebSocketDisconnect:
        # Client disconnected; the generator will be cancelled and listener removed
        pass
