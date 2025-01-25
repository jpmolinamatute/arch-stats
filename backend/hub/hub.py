import asyncio
from logging import Logger
from multiprocessing import Process
from os import getenv
from uuid import UUID

import asyncpg

# from psycopg.types.json import Json
from websockets import ConnectionClosed, Subprotocol
from websockets.asyncio.server import ServerConnection, serve


# Channel to listen to in PostgreSQL
LISTEN_CHANNEL = "shooting_change"

# Global set of connected WebSocket clients
CONNECTED_CLIENTS: set[ServerConnection] = set()
CLIENTS_LOCK = asyncio.Lock()


async def notify_clients(message: str) -> None:
    """Send a message to all connected WebSocket clients."""
    async with CLIENTS_LOCK:
        for client in CONNECTED_CLIENTS:
            await client.send(message)


async def create_temporary_trigger(conn: asyncpg.Connection, archer_id: UUID) -> None:

    trigger_creation = f"""
    CREATE TEMP TRIGGER shooting_change_trigger
    AFTER INSERT OR UPDATE ON shooting
    FOR EACH ROW
    EXECUTE FUNCTION notify_shooting_change({str(archer_id)});
    """

    await conn.execute(trigger_creation)


async def listen_to_db(archer_id: UUID) -> None:
    """Listen for notifications from PostgreSQL."""
    user = getenv("ARCH_STATS_USER")
    listen_channel = "shooting_change"
    params = {}
    if user:
        params["user"] = user
    conn = await asyncpg.connect(**params)
    try:
        await create_temporary_trigger(conn, archer_id)
        await conn.add_listener(
            listen_channel, lambda *args: asyncio.create_task(notify_clients(args[-1]))
        )
        print(f"Listening on channel '{listen_channel}'...")

        while True:
            await asyncio.sleep(1)  # Keep the connection alive
    finally:
        await conn.close()


def websocket_client_handler(archer_id: UUID) -> None:
    """Handle WebSocket connections in a separate process."""
    asyncio.run(listen_to_db(archer_id))


async def websocket_handler(websocket: ServerConnection) -> None:
    """Handle WebSocket connections."""
    print("New WebSocket client connected.")
    CONNECTED_CLIENTS.add(websocket)
    process: Process | None = None
    try:
        archer_id_str = await websocket.recv()
        if isinstance(archer_id_str, str):  # Receive the archer_id from the client
            archer_id = UUID(archer_id_str)
            process = Process(target=websocket_client_handler, args=(archer_id,))
            process.start()
            await websocket.wait_closed()
        else:
            print("Invalid archer_id received.")
    except ConnectionClosed:
        print("WebSocket connection closed.")
    finally:
        CONNECTED_CLIENTS.remove(websocket)
        if process:
            process.terminate()
        print("WebSocket client disconnected.")


async def setup(logger: Logger) -> None:
    # Run the WebSocket server
    logger.info("Starting the hub...")
    websocket_port = int(getenv("ARCH_STATS_WS_PORT", "8765"))
    websocket_host = getenv("ARCH_STATS_HOSTNAME", "localhost")
    async with serve(
        websocket_handler, websocket_host, websocket_port, subprotocols=[Subprotocol("shooting")]
    ):
        logger.info(f"WebSocket server running on ws://{websocket_host}:{websocket_port}")
        await asyncio.Future()  # Run forever
