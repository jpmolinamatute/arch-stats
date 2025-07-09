import json
from asyncio import CancelledError, sleep

from asyncpg import Pool
from asyncpg.connection import Connection
from fastapi import APIRouter, Depends, WebSocket, WebSocketDisconnect

from server.db_pool import DBPool
from server.settings import settings


WSRouter = APIRouter()


async def get_db_pool() -> Pool:
    return await DBPool.get_db_pool()


@WSRouter.websocket("/ws/shot")
async def websocket_shot(websocket: WebSocket, db_pool: Pool = Depends(get_db_pool)) -> None:
    """WebSocket endpoint for real-time communication."""
    await websocket.accept()
    should_run = True
    channel = settings.arch_stats_ws_channel

    async def handle_notification(_: Connection, __: int, ___: str, payload: str) -> None:
        # Parse payload and send via websocket
        nonlocal should_run
        try:
            data = json.loads(payload)
        except Exception:
            data = payload
        if should_run:
            try:
                await websocket.send_json(data)
            except (WebSocketDisconnect, RuntimeError, KeyboardInterrupt):
                should_run = False

    async with db_pool.acquire() as conn:
        try:
            await conn.add_listener(channel, handle_notification)
            # Keep the websocket open until client disconnects
            while should_run:
                await sleep(1)
        except CancelledError:
            should_run = False
            raise
        finally:
            await conn.remove_listener(channel, handle_notification)


# @WSRouter.websocket("/ws/arrow")
# async def websocket_arrow(
#     websocket: WebSocket,
#     db_pool: Pool = Depends(get_db_pool),
#     channel: str = "new_arrow",
# ) -> None:
#     """WebSocket endpoint for real-time communication."""
#     await websocket.accept()
