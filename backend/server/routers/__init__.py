from server.routers.archer import router as router_archer
from server.routers.lane import router as router_lane
from server.routers.shooting import router as router_shooting
from server.routers.target import router as router_target
from server.routers.tournament import router as router_tournament

__all__ = [
    "router_archer",
    "router_lane",
    "router_shooting",
    "router_target",
    "router_tournament",
]
