import sys
import asyncio
from os import getenv

import psycopg
from psycopg import sql

# from psycopg.types.json import Json
from websockets import ConnectionClosed, Subprotocol
from websockets.asyncio.server import ServerConnection, serve

# WebSocket server details
WEBSOCKET_HOST = "localhost"
WEBSOCKET_PORT = 8765

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


async def listen_to_db() -> None:
    """Listen for notifications from PostgreSQL."""
    user = getenv("ARCH_STATS_USER")
    params = {}
    if user:
        params["user"] = user
    async with await psycopg.AsyncConnection.connect(**params) as conn:  # type: ignore[arg-type]
        # Ensure the channel is being listened to
        async with conn.cursor() as cur:
            await cur.execute(
                sql.SQL("LISTEN {channel};").format(channel=sql.Identifier(LISTEN_CHANNEL))
            )
            await conn.commit()
            print(f"Listening on channel '{LISTEN_CHANNEL}'...")

            while True:
                # Wait for notifications
                async for notification in conn.notifies():
                    print(f"Notification received: {notification.payload}")
                    # Forward notification to all connected WebSocket clients
                    await notify_clients(notification.payload)


async def websocket_handler(websocket: ServerConnection) -> None:
    """Handle WebSocket connections."""
    print("New WebSocket client connected.")
    CONNECTED_CLIENTS.add(websocket)
    try:
        await websocket.wait_closed()
    except ConnectionClosed:
        print("WebSocket connection closed.")
    finally:
        CONNECTED_CLIENTS.remove(websocket)
        print("WebSocket client disconnected.")


async def main() -> None:
    # Run the WebSocket server
    websocket_server = serve(
        websocket_handler, WEBSOCKET_HOST, WEBSOCKET_PORT, subprotocols=[Subprotocol("shooting")]
    )
    print(f"WebSocket server running on ws://{WEBSOCKET_HOST}:{WEBSOCKET_PORT}")

    # Run the PostgreSQL listener and WebSocket server concurrently
    await asyncio.gather(listen_to_db(), websocket_server)


if __name__ == "__main__":
    EXIT_STATUS = 0
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("Bye!")
    except Exception:
        print("An unexpected error occurred")
        EXIT_STATUS = 1
    finally:
        sys.exit(EXIT_STATUS)
