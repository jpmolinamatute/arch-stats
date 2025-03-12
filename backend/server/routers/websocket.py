import json
from asyncio import CancelledError, Lock, create_task, sleep
from collections import defaultdict
from typing import Any, MutableMapping
from uuid import UUID

from asyncpg import Connection, Pool
from fastapi import APIRouter, Depends, WebSocket, WebSocketDisconnect

from server.dependencies import get_db
from shared.logger import get_logger


router = APIRouter()
Message = MutableMapping[str, Any]
RATE_LIMIT_INTERVAL = 0.5  # 0.5 seconds between messages
MAX_MESSAGE_SIZE = 1024  # 1 KB maximum message size


async def pg_listener(_: Connection, pid: int, channel: str, payload: str) -> None:
    """
    Called automatically when Postgres triggers a NOTIFY on `channel`.
    `payload` is a JSON string from `row_to_json(NEW)`.
    """
    # Convert string to Python dict if you want to manipulate. For now, just forward it.
    # But let's parse to a dict so we can broadcast as JSON.

    try:
        data = json.loads(payload)
    except json.JSONDecodeError:
        data = {"raw": payload}

    archer_id = data.get("archer_id")
    if not archer_id:
        # If no archer_id is found in row, do nothing or fallback
        return

    await manager.send_to_archer(archer_id, {"type": "pg_notify", "payload": data})


def get_archer_uuid(archer_id: str) -> UUID | None:
    """
    Check & return a valid UUID4 from archer_id.
    """
    valid = None
    try:
        valid = UUID(archer_id, version=4)
    except ValueError:
        pass
    return valid


class ConnectionManager:
    """Manages WebSocket connections."""

    def __init__(self) -> None:
        self.archer_connections: dict[UUID, set[WebSocket]] = defaultdict(set)
        self.lock = Lock()
        self.logger = get_logger(__name__)

    async def connect(self, websocket: WebSocket, archer_id: UUID) -> None:
        """Accepts new WebSocket connection."""
        async with self.lock:
            self.logger.info("New WebSocket connection established.")
            self.archer_connections[archer_id].add(websocket)

    async def disconnect(self, websocket: WebSocket, msg: str, archer_id: UUID) -> None:
        """Safely removes a disconnected WebSocket connection."""
        async with self.lock:
            if (
                archer_id in self.archer_connections
                and websocket in self.archer_connections[archer_id]
            ):
                self.logger.info("WebSocket connection closed. %s", msg)
                self.archer_connections[archer_id].remove(websocket)
                if not self.archer_connections[archer_id]:
                    del self.archer_connections[archer_id]

    async def broadcast(self, message: Message) -> None:
        """
        Broadcast a JSON message to all active connections.
        If a socket is closed, remove it from the active list.
        """
        async with self.lock:
            for archer_id in list(self.archer_connections.keys()):
                for conn in self.archer_connections[archer_id]:
                    try:
                        await conn.send_json(message)
                    except Exception as exc:
                        await self.disconnect(
                            conn, f"Send failed; removing connection: {str(exc)}", archer_id
                        )

    async def heart_beep(self) -> None:
        """Periodically check/removes stale WebSocket connections."""

        await self.broadcast({"type": "ping", "payload": "health_check"})

    def get_partial_archer_id(self, archer_id: UUID) -> str:
        """Extracts the partial archer_id from the full archer_id."""
        archer_id_str = str(archer_id)
        return archer_id_str.split("-", maxsplit=1)[0]

    async def create_temporary_notify(
        self,
        conn: Connection,
        channel: str,
        archer_id: UUID,
    ) -> None:
        """Create a temporary notify function that triggers
        `pg_notify(channel, row_to_json(NEW)::text)` upon insert/update."""

        self.logger.info("Creating temporary notify_%s function...", channel)
        notify_creation = f"""
            CREATE OR REPLACE FUNCTION notify_{channel}()
            RETURNS TRIGGER AS $$
            DECLARE
                    _row jsonb;
            BEGIN
                RAISE NOTICE 'The function notify_{channel}() was called';
                IF EXISTS (
                    SELECT 1
                    FROM arrow
                    WHERE id = NEW.arrow_id
                    AND archer_id = '$1'
                ) THEN
                    _row = jsonb_build_object(
                        'id', NEW.id,
                        'archer_id', $1,
                        'arrow_id', NEW.arrow_id,
                        'target_track_id', NEW.target_track_id,
                        'arrow_engage_time', NEW.arrow_engage_time,
                        'draw_length', NEW.draw_length,
                        'arrow_disengage_time', NEW.arrow_disengage_time,
                        'arrow_landing_time', NEW.arrow_landing_time,
                        'distance', NEW.distance,
                        'x_coordinate', NEW.x_coordinate,
                        'y_coordinate', NEW.y_coordinate
                    );
                    PERFORM pg_notify('$2', row_to_json(_row)::text);
                END IF;
                RETURN NEW;
            END;
            $$ LANGUAGE plpgsql;
        """
        await conn.execute(notify_creation, archer_id, channel)

    async def create_temporary_trigger(self, conn: Connection, channel: str) -> None:
        """Create a temporary trigger for a specific archer."""

        self.logger.info("Creating temporary trigger_%s trigger...", channel)
        trigger_creation = f"""
        CREATE TRIGGER trigger_{channel}
        AFTER INSERT OR UPDATE ON shooting
        FOR EACH ROW
        EXECUTE FUNCTION notify_{channel}();
        """

        await conn.execute(trigger_creation)

    async def drop_temporary_trigger(self, conn: Connection, channel: str) -> None:
        """Drop a temporary trigger for a specific archer."""

        self.logger.info("Dropping temporary trigger_%s trigger...", channel)
        trigger_deletion = f"DROP TRIGGER trigger_{channel} ON shooting;"

        await conn.execute(trigger_deletion)

    async def drop_temporary_notify(self, conn: Connection, channel: str) -> None:
        """Drop a temporary notify function."""

        self.logger.info("Dropping temporary notify_%s function...", channel)
        notify_deletion = f"DROP FUNCTION notify_{channel}();"
        await conn.execute(notify_deletion)

    async def send_to_archer(self, archer_id: UUID, message: dict[str, Any]) -> None:
        """
        Send `message` to all WebSockets belonging to the given archer_id.
        """
        if archer_id not in self.archer_connections:
            return  # No active connections for that archer

        for ws in self.archer_connections[archer_id]:
            try:
                await ws.send_json(message)
            except Exception as exc:
                await self.disconnect(ws, f"Send failed: {str(exc)}", archer_id)


manager = ConnectionManager()


async def keep_websocket_alive(interval_seconds: float = 30.0) -> None:
    """
    Runs in the background, periodically calls `manager.heart_beep()`
    to keep all active WebSockets alive. Cancels when parent task ends.
    """
    try:
        while True:
            await sleep(interval_seconds)
            await manager.heart_beep()
    except CancelledError:
        # Normal behavior on disconnect or teardown
        manager.logger.info("Keep-alive task canceled.")
        raise


@router.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket, db_pool: Pool = Depends(get_db)) -> None:
    """WebSocket endpoint for real-time communication."""
    await websocket.accept()

    async with db_pool.acquire() as conn:
        keepalive_task = None
        channel = None
        archer_id = None
        try:
            while True:
                # Receive text from WebSocket
                client_msg = await websocket.receive_text()

                # Validate input size
                if len(client_msg) > MAX_MESSAGE_SIZE:
                    await websocket.send_json({"error": "Message size exceeds limit."})
                    continue

                archer_id = get_archer_uuid(client_msg)
                if not archer_id:
                    await websocket.send_json({"error": "Invalid archer ID format."})
                    continue

                await manager.connect(websocket, archer_id)
                if channel is None:
                    # Only do this once per connection if needed
                    partial_uuid = manager.get_partial_archer_id(archer_id)
                    channel = f"listener_{partial_uuid}"

                    await manager.create_temporary_notify(conn, channel, archer_id)
                    await manager.create_temporary_trigger(conn, channel)
                    await conn.add_listener(channel, pg_listener)
                    manager.logger.info("Started listening on channel '%s'", channel)
                    keepalive_task = create_task(keep_websocket_alive())

                await sleep(RATE_LIMIT_INTERVAL)

        except WebSocketDisconnect:
            if isinstance(archer_id, UUID):
                manager.logger.info("WebSocketDisconnect caught in /ws")
                await manager.disconnect(websocket, "Client disconnected.", archer_id)
        except CancelledError:
            if isinstance(archer_id, UUID):
                manager.logger.info("Task canceled in /ws endpoint")
                await manager.disconnect(websocket, "Task canceled.", archer_id)
        except Exception as exc:
            if isinstance(archer_id, UUID):
                manager.logger.error("Unexpected error: %s", exc)
                await manager.disconnect(websocket, f"Unexpected error: {exc}", archer_id)
        finally:
            # 5) Cleanup: remove listener, drop triggers/functions
            if channel:
                manager.logger.info("Removing listener on channel '%s'...", channel)
                await conn.remove_listener(channel, pg_listener)
                await manager.drop_temporary_trigger(conn, channel)
                await manager.drop_temporary_notify(conn, channel)
                manager.logger.info("Stopped listening on channel '%s'.", channel)
            if keepalive_task:
                keepalive_task.cancel()
