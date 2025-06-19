import logging
from asyncio import CancelledError
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from pathlib import Path

from asyncpg import Pool
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from server.models import ArrowsDB, DBState, SessionsDB, ShotsDB, TargetsDB
from server.routers import ArrowsRouter, SessionsRouter, ShotsRouter, TargetsRouter, WSRouter
from server.settings import settings


async def create_tables(pool: Pool) -> None:
    arrows = ArrowsDB(pool)
    shots = ShotsDB(pool)
    sessions = SessionsDB(pool)
    targets = TargetsDB(pool)
    channel = settings.arch_stats_ws_channel
    await arrows.create_table()
    await sessions.create_table()
    await shots.create_table()
    await shots.create_notification(channel)
    await targets.create_validation_function()
    await targets.create_table()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Startup and shutdown logic for the app."""
    logger: logging.Logger = getattr(app.state, "logger")
    await DBState.init_db()  # Initialize database
    pool = await DBState.get_db_pool()
    await create_tables(pool)
    try:
        yield
    except CancelledError:
        logger.info("Shutdown interrupted. Cleaning up...")
    finally:
        await DBState.close_db()  # Close database safely


def run() -> FastAPI:
    logger = logging.getLogger("arch-stats")
    server_name = settings.postgres_host
    server_port = settings.arch_stats_server_port
    dev_mode = settings.arch_stats_dev_mode
    version = "0.1.0"
    if dev_mode:
        logger.info("Starting the server on %s:%d", server_name, server_port)
    else:
        logger.info("Starting the server in production mode on %s:%d", server_name, server_port)
    app = FastAPI(
        debug=dev_mode,
        lifespan=lifespan,
        openapi_url="/api/openapi.json",
        title="Arch Stats API",
        version=version,
    )
    mayor_version = f"v{version[0]}"
    # Include blueprints
    app.state.logger = logger
    app.include_router(SessionsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(ArrowsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(ShotsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(TargetsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(WSRouter, prefix=f"/api/{mayor_version}")
    current_file_path = Path(__file__).parent
    app.mount(
        "/app",
        StaticFiles(
            directory=current_file_path.joinpath("frontend"),
            html=True,
            check_dir=True,
            follow_symlink=True,
        ),
        name="frontend",
    )

    return app


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)

    run()
