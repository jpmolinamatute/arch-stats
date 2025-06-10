import logging
from asyncio import CancelledError
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from os import getenv
from pathlib import Path

from asyncpg import Pool
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from server.models import ArrowsDB, DBState, SessionsDB, ShotsDB, TargetsDB
from server.routers import ArrowsRouter, SessionsRouter, ShotsRouter, TargetsRouter, WSRouter


async def create_tables(pool: Pool) -> None:
    arrows = ArrowsDB(pool)
    shots = ShotsDB(pool)
    sessions = SessionsDB(pool)
    targets = TargetsDB(pool)
    await arrows.create_table()
    await sessions.create_table()
    await shots.create_table()
    await shots.create_notification("archy")
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
    server_name = getenv("ARCH_STATS_HOSTNAME", "localhost")
    server_port = int(getenv("ARCH_STATS_SERVER_PORT", "8000"))
    dev_mode = getenv("ARCH_STATS_DEV_MODE", "False").lower() == "true"
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
        "/",
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
