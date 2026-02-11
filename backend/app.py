from asyncio import CancelledError
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, status
from fastapi.staticfiles import StaticFiles

from core import DBPool, get_logger, settings
from routers.v0 import (
    archer_router,
    auth_router,
    faces_router,
    session_router,
    shot_router,
    slot_router,
    websocket,
)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None]:
    """Startup and shutdown logic for the app."""

    try:
        app.state.logger = get_logger()
        app.state.logger.info("Starting Server up...")
        app.state.db_pool = await DBPool.open_db_pool()
        yield
    except CancelledError:
        app.state.logger.error("Shutdown interrupted. Cleaning up...")
    finally:
        app.state.logger.debug("Closing DB...")
        await DBPool.close_db_pool()
        app.state.logger.info("Server shutdown complete.")


def run() -> FastAPI:
    version = "0.2.0"
    current_file_path = Path(__file__).parent
    mayor_version = f"v{version[0]}"
    dev_mode = settings.arch_stats_dev_mode
    app = FastAPI(
        debug=dev_mode,
        lifespan=lifespan,
        openapi_url="/api/openapi.json",
        docs_url="/api/swagger",
        title="Arch Stats API",
        version=version,
        openapi_tags=[
            {"name": "Archers", "description": "Operations about archers domain"},
            {"name": "Auth", "description": "Authentication endpoints"},
            {"name": "Sessions", "description": "Operations about sessions domain"},
            {"name": "Slots", "description": "Operations about slot assignments"},
            {"name": "Shots", "description": "Operations about shots domain"},
            {"name": "Faces", "description": "Operations about target faces domain"},
            {"name": "Websocket", "description": "Websocket endpoints"},
        ],
    )

    # Routers
    app.include_router(auth_router, prefix=f"/api/{mayor_version}")
    app.include_router(archer_router, prefix=f"/api/{mayor_version}")
    app.include_router(session_router, prefix=f"/api/{mayor_version}")
    app.include_router(slot_router, prefix=f"/api/{mayor_version}")
    app.include_router(shot_router, prefix=f"/api/{mayor_version}")
    app.include_router(faces_router, prefix=f"/api/{mayor_version}")
    app.include_router(websocket, prefix=f"/api/{mayor_version}")

    @app.api_route(
        "/api/{path:path}",
        methods=["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"],
        status_code=status.HTTP_404_NOT_FOUND,
        include_in_schema=False,
    )
    async def api_not_found(path: str) -> dict[str, str]:
        return {"detail": f"{path} Not Found"}

    app.mount(
        "",
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
    run()
