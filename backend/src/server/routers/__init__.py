from server.routers.arrows_router import ArrowsRouter
from server.routers.faces_router import FacesRouter
from server.routers.session_performance_router import SessionPerformanceRouter
from server.routers.sessions_router import SessionsRouter
from server.routers.shots_router import ShotsRouter
from server.routers.targets_router import TargetsRouter
from server.routers.websocket import WSRouter


__all__ = [
    "ArrowsRouter",
    "FacesRouter",
    "ShotsRouter",
    "SessionsRouter",
    "TargetsRouter",
    "WSRouter",
    "SessionPerformanceRouter",
]
