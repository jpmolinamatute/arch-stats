from fastapi import FastAPI
import uvicorn
from server.routers import (
    router_archer,
    router_lane,
    router_shooting,
    router_target,
    router_tournament,
)


def create_app() -> FastAPI:
    app = FastAPI()

    # Include blueprints
    app.include_router(router_tournament, prefix="/api")
    app.include_router(router_archer, prefix="/api")
    app.include_router(router_lane, prefix="/api")
    app.include_router(router_shooting, prefix="/api")
    app.include_router(router_target, prefix="/api")

    @app.get("/", response_model=dict)
    async def root() -> dict[str, str]:
        # this route will serve the static files (HTML, JS, CSS) aka the frontend
        return {"message": "Welcome to the archery management system!"}

    return app


if __name__ == "__main__":
    APP = create_app()
    uvicorn.run(APP, host="0.0.0.0", port=8000)
