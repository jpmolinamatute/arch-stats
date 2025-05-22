import logging
from asyncio import CancelledError
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from os import getenv
from pathlib import Path

import uvicorn
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from server.models import ArrowsDB, DBState, SessionsDB, ShotsDB, TargetsDB
from server.routers import ArrowsRouter, SessionsRouter, ShotsRouter, TargetsRouter, ArrowUUIDRouter


async def create_tables() -> None:
    pool = await DBState.get_db_pool()
    arrows = ArrowsDB(pool)
    shots = ShotsDB(pool)
    sessions = SessionsDB(pool)
    targets = TargetsDB(pool)
    await arrows.create_table()
    await sessions.create_table()
    await shots.create_table()
    await targets.create_table()


@asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncGenerator[None, None]:
    """Startup and shutdown logic for the app."""
    await DBState.init_db()  # Initialize database
    await create_tables()
    try:
        yield
    except CancelledError:
        print("Shutdown interrupted. Cleaning up...")
    finally:
        await DBState.close_db()  # Close database safely


def create_app() -> FastAPI:
    version = "0.1.0"
    app = FastAPI(
        lifespan=lifespan,
        openapi_url="/api/openapi.json",
        title="Arch Stats API",
        version=version,
    )
    mayor_version = f"v{version[0]}"
    # Include blueprints
    app.include_router(ArrowsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(ShotsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(SessionsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(TargetsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(ArrowUUIDRouter, prefix=f"/api/{mayor_version}")
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

    if dev_mode:
        worker = None
        color = True
    else:
        worker = 4
        color = None
    config = uvicorn.Config(
        app=create_app(),
        host=server_name,
        port=server_port,
        loop="uvloop",
        lifespan="on",
        reload=dev_mode,
        ws="websockets",
        http="h11",
        workers=worker,
        use_colors=color,
    )
    server = uvicorn.Server(config)

    await server.serve()
