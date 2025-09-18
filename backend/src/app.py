import logging
from asyncio import CancelledError
from collections.abc import AsyncGenerator, Awaitable, Callable
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from starlette.middleware.base import BaseHTTPMiddleware

from core import DBPool, LoggerFactory, settings
from routers.v0 import archer_router, auth_router, session_router


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Startup and shutdown logic for the app."""

    try:
        app.state.logger = LoggerFactory().get_logger("server")
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
    app = FastAPI(
        debug=settings.arch_stats_dev_mode,
        lifespan=lifespan,
        openapi_url="/api/openapi.json",
        docs_url="/api/swagger",
        title="Arch Stats API",
        version=version,
        openapi_tags=[
            {"name": "Archers", "description": "Operations about archers domain"},
            {"name": "Auth", "description": "Authentication endpoints"},
            {"name": "Sessions", "description": "Operations about sessions domain"},
        ],
    )

    class ReferrerPolicyMiddleware(BaseHTTPMiddleware):
        async def dispatch(
            self, request: Request, call_next: Callable[[Request], Awaitable[Response]]
        ) -> Response:
            response: Response = await call_next(request)
            response.headers.setdefault("Referrer-Policy", "no-referrer-when-downgrade")
            # Dev-friendly security headers to avoid COOP/COEP blocking postMessage
            # Only set in dev mode (default here is permissive; tighten for prod behind a proxy)
            if settings.arch_stats_dev_mode:
                response.headers.setdefault("Cross-Origin-Opener-Policy", "unsafe-none")
                response.headers.setdefault("Cross-Origin-Embedder-Policy", "unsafe-none")
                response.headers.setdefault("Permissions-Policy", "identity-credentials-get=(*)")
            return response

    app.add_middleware(ReferrerPolicyMiddleware)

    # Basic CORS for local frontend dev (http://localhost:5173)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=[
            "http://localhost:5173",
            "http://127.0.0.1:5173",
        ],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Routers
    app.include_router(auth_router, prefix=f"/api/{mayor_version}")
    app.include_router(archer_router, prefix=f"/api/{mayor_version}")
    app.include_router(session_router, prefix=f"/api/{mayor_version}")
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
    logging.basicConfig(level=logging.DEBUG)
    run()
