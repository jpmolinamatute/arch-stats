from routers.v0.archer_router import router as archer_router
from routers.v0.auth_router import router as auth_router
from routers.v0.session_router import router as session_router
from routers.v0.shot_router import router as shot_router
from routers.v0.slot_router import router as slot_router


__all__ = ["auth_router", "archer_router", "session_router", "slot_router", "shot_router"]
