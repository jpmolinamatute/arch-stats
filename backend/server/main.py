import logging
from asyncio import CancelledError
from contextlib import asynccontextmanager
from os import getenv
from pathlib import Path
from collections.abc import AsyncGenerator

import uvicorn
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from database.db_state import DBState  # Import database functions
from server.routers import arrow, session, shot


@asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncGenerator[None, None]:
    """Startup and shutdown logic for the app."""
    await DBState.init_db()  # Initialize database
    try:
        yield
    except CancelledError:
        print("Shutdown interrupted. Cleaning up...")
    finally:
        await DBState.close_db()  # Close database safely


def create_app() -> FastAPI:
    app = FastAPI(lifespan=lifespan, openapi_url="/api/openapi.json", title="Arch Stats API")

    # Include blueprints
    app.include_router(arrow, prefix="/api")
    app.include_router(shot, prefix="/api")
    app.include_router(session, prefix="/api")
    # app.include_router(router_shooting, prefix="/api")
    # app.include_router(router_target, prefix="/api")
    # app.include_router(router_websocket)
    current_file_path = Path(__file__).parent
    frontend_path = current_file_path.joinpath("frontend")
    app.mount(
        "/ui/", StaticFiles(directory=frontend_path, html=True, check_dir=True), name="frontend"
    )

    return app


async def setup(logger: logging.Logger) -> None:
    server_name = getenv("ARCH_STATS_HOSTNAME", "localhost")
    server_port = int(getenv("ARCH_STATS_SERVER_PORT", "8000"))
    dev_mode = getenv("ARCH_STATS_DEV_MODE", "False").lower() == "true"
    logger.info("Starting the server on %s:%d", server_name, server_port)
    logger.info("Development mode: %s", dev_mode)
    app = create_app()
    config = uvicorn.Config(
        app,
        host=server_name,
        port=server_port,
        loop="uvloop",
        lifespan="on",
        reload=dev_mode,
        ws="websockets",
        http="h11",
        workers=4,
    )
    server = uvicorn.Server(config)
    await server.serve()
