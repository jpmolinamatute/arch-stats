from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, WebSocket, WebSocketDisconnect, status

from core import LiveStatsManager
from routers.deps.auth import require_auth
from routers.deps.models import get_live_stats_manager, get_live_stats_manager_ws
from schema import WebSocketMessage, WSContentType
from schema.live_stats_schema import LiveStat

router = APIRouter(prefix="/stats", tags=["Stats"])


@router.get("/{slot_id:uuid}", response_model=LiveStat, status_code=status.HTTP_200_OK)
async def get_stats(
    slot_id: UUID,
    current_archer_id: Annotated[UUID, Depends(require_auth)],
    live_stats_manager: Annotated[LiveStatsManager, Depends(get_live_stats_manager)],
) -> LiveStat:
    """
    Get live statistics and shots for a slot.
    """
    return await live_stats_manager.get_stats(slot_id, current_archer_id)


@router.websocket("/ws/{slot_id:uuid}")
async def websocket_stats(
    websocket: WebSocket,
    slot_id: UUID,
    live_stats_manager: Annotated[LiveStatsManager, Depends(get_live_stats_manager_ws)],
) -> None:
    """WebSocket endpoint streaming shot notifications for a slot.

    Relays DB NOTIFY payloads from channel "shot_insert_{slot_id}" to the client
    as JSON when possible; falls back to raw text otherwise.
    """

    await websocket.accept()
    try:
        async for payload in live_stats_manager.listen_for_shots(slot_id):
            message = WebSocketMessage(content=payload, content_type=WSContentType.SHOT_CREATED)
            await websocket.send_json(message.model_dump(mode="json"))
    except WebSocketDisconnect:
        # Client disconnected; the generator will be cancelled and listener removed
        pass
