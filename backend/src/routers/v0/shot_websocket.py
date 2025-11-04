from uuid import UUID

from fastapi import APIRouter, Depends, WebSocket, WebSocketDisconnect

from models import ShotModel
from routers.deps.models import get_shot_model


router = APIRouter()


@router.websocket("/ws/shot/by-slot/{slot_id:uuid}")
async def websocket_shot(
    websocket: WebSocket,
    slot_id: UUID,
    shot_model: ShotModel = Depends(get_shot_model),
) -> None:
    """WebSocket endpoint streaming shot notifications for a slot.

    Relays DB NOTIFY payloads from channel "shot_insert_{slot_id}" to the client
    as JSON when possible; falls back to raw text otherwise.
    """

    await websocket.accept()
    try:
        async for payload in shot_model.listen_for_shots(slot_id):
            await websocket.send_json(payload)
    except WebSocketDisconnect:
        # Client disconnected; the generator will be cancelled and listener removed
        pass
