import logging
from asyncio import CancelledError
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from pathlib import Path

from asyncpg import Pool
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from typing_extensions import Literal

from server.routers import ArrowsRouter, SessionsRouter, ShotsRouter, TargetsRouter, WSRouter
from shared.db_pool import DBPool
from shared.logger import LogLevel, get_logger
from shared.models import ArrowsModel, FacesModel, SessionsModel, ShotsModel, TargetsModel
from shared.settings import settings


TablesAction = Literal["drop", "create"]


async def manage_tables(pool: Pool, action: TablesAction) -> None:
    arrows = ArrowsModel(pool)
    shots = ShotsModel(pool)
    sessions = SessionsModel(pool)
    targets = TargetsModel(pool)
    faces = FacesModel(pool)

    if action == "create":
        await arrows.create()
        await sessions.create()
        await shots.create()
        await targets.create()
        await faces.create()
    elif action == "drop":
        # Drop child/dependent tables first to satisfy FK constraints
        await faces.drop()
        await targets.drop()
        await shots.drop()
        await arrows.drop()
        await sessions.drop()


def log_start(logger: logging.Logger, dev_mode: bool) -> None:
    if dev_mode:
        logger.info("Starting the server in development mode")
    else:
        logger.info("Starting the server")


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Startup and shutdown logic for the app."""

    try:
        log_level = LogLevel.DEBUG if settings.arch_stats_dev_mode else LogLevel.INFO
        app.state.logger = get_logger("server", log_level)
        app.state.db_pool = await DBPool.open_db_pool()
        log_start(app.state.logger, app.debug)
        app.state.logger.debug("Initializing DB...")
        await manage_tables(app.state.db_pool, "create")
        app.state.logger.debug("DB initialized.")
        yield
    except CancelledError:
        app.state.logger.info("Shutdown interrupted. Cleaning up...")
        if app.debug:
            await manage_tables(app.state.db_pool, "drop")
    finally:
        app.state.logger.debug("Closing DB...")
        await DBPool.close_db_pool()
        app.state.logger.debug("DB closed.")


def run() -> FastAPI:
    version = "0.1.0"
    current_file_path = Path(__file__).parent
    mayor_version = f"v{version[0]}"
    app = FastAPI(
        debug=settings.arch_stats_dev_mode,
        lifespan=lifespan,
        openapi_url="/api/openapi.json",
        docs_url="/api/swagger",
        title="Arch Stats API",
        version=version,
    )

    app.include_router(SessionsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(ArrowsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(ShotsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(TargetsRouter, prefix=f"/api/{mayor_version}")
    app.include_router(WSRouter, prefix=f"/api/{mayor_version}")
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
