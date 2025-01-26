from asyncio import Lock, sleep
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


async def pg_listener(_: Connection, pid: int, channel: str, message: Message) -> None:
    """Handles notifications from PostgreSQL."""
    print(f"pid: {pid}, channel: {channel}")
    await manager.broadcast(message)


class ConnectionManager:
    """Manages WebSocket connections."""

    def __init__(self) -> None:
        self.active_connections: list[WebSocket] = []
        self.lock = Lock()
        self.logger = get_logger(__name__)

    async def connect(self, websocket: WebSocket) -> None:
        """Accepts new WebSocket connection."""
        await websocket.accept()
        async with self.lock:
            self.logger.info("New WebSocket connection established.")
            self.active_connections.append(websocket)

    def get_active_connections(self) -> list[WebSocket]:
        """Returns all active WebSocket connections."""
        return self.active_connections

    async def disconnect(self, websocket: WebSocket) -> None:
        """Safely removes a disconnected WebSocket connection."""
        async with self.lock:
            if websocket in self.active_connections:
                self.logger.info("WebSocket connection closed.")
                self.active_connections.remove(websocket)

    async def broadcast(self, message: Message) -> None:
        """Sends a message to all connected clients."""
        async with self.lock:
            for connection in self.active_connections:
                try:
                    self.logger.info("Broadcasting message: %s", message)
                    await connection.send_json(message)
                except Exception as e:
                    self.logger.error("Error broadcasting to WebSocket: %s", e)
                    await self.disconnect(connection)

    async def prune_inactive_connections(self) -> None:
        """Periodically check/removes stale WebSocket connections."""

        await self.broadcast({"ping": "health_check"})

    def get_partial_archer_id(self, archer_id: str) -> str:
        """Extracts the partial archer_id from the full archer_id."""
        return archer_id.split("-")[0]

    async def create_temporary_notify(
        self,
        conn: Connection,
        channel: str,
        archer_id: str,
    ) -> None:
        """Create a temporary notify function."""

        self.logger.info("Creating temporary notify_%s function...", channel)
        notify_creation = f"""
            CREATE OR REPLACE FUNCTION notify_{channel}()
            RETURNS TRIGGER AS $$
            BEGIN
                RAISE NOTICE 'The function notify_{channel}() was called';
                IF EXISTS (
                    SELECT 1
                    FROM arrow
                    WHERE id = NEW.arrow_id
                    AND archer_id = '{archer_id}'
                ) THEN
                    PERFORM pg_notify('{channel}', row_to_json(NEW)::text);
                END IF;
                RETURN NEW;
            END;
            $$ LANGUAGE plpgsql;
        """
        await conn.execute(notify_creation)

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

    async def keep_conn_alive(self) -> None:
        """Keeps the connection alive."""
        while True:
            await sleep(30)  # Adjust interval as needed
            await self.prune_inactive_connections()

    def validate_archer_id(self, archer_id: str) -> bool:
        """Validates the archer_id."""
        valid = True
        try:
            self.logger.info("Validating archer_id: %s", archer_id)
            UUID(archer_id, version=4)
        except ValueError:
            valid = False
        return valid

    async def listen(self, archer_id: str, conn: Connection) -> None:
        self.logger.info("Listening for archer_id: %s", archer_id)
        partial_uuid = self.get_partial_archer_id(archer_id)
        channel = f"listener_{partial_uuid}"

        try:
            await self.create_temporary_notify(conn, channel, archer_id)
            await self.create_temporary_trigger(conn, channel)
            await conn.add_listener(channel, pg_listener)
            self.logger.info("Listening on channel '%s'...", channel)
            await self.keep_conn_alive()
        except Exception as e:
            self.logger.error("Error: %s", e)
        finally:
            self.logger.info("Removing listener on channel '%s'...", channel)
            await conn.remove_listener(channel, pg_listener)
            await self.drop_temporary_trigger(conn, channel)
            await self.drop_temporary_notify(conn, channel)
            self.logger.info("Stopped listening on channel '%s'.", channel)


manager = ConnectionManager()


@router.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket, db_pool: Pool = Depends(get_db)) -> None:
    """WebSocket endpoint for real-time communication."""

    async with db_pool.acquire() as conn:
        await manager.connect(websocket)
        try:
            while True:
                # Receive text from WebSocket
                archer_id = await websocket.receive_text()

                # Validate input size
                if len(archer_id) > MAX_MESSAGE_SIZE:
                    await websocket.send_json({"error": "Message size exceeds limit."})
                    continue

                # Validate Archer ID
                if not manager.validate_archer_id(archer_id):
                    await websocket.send_json({"error": "Invalid archer ID format."})
                    continue

                # Handle listening logic
                await manager.listen(archer_id, conn)
                await sleep(RATE_LIMIT_INTERVAL)
        except WebSocketDisconnect:
            await manager.disconnect(websocket)
