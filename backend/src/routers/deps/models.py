import logging

from asyncpg import Pool
from fastapi import Request, WebSocket

from core import SlotManager
from models import ArcherModel, SessionModel, ShotModel, SlotModel
from routers.deps.auth import require_auth


async def get_session_model(request: Request) -> SessionModel:
    """Dependency provider returning a `SessionModel` bound to the pool.

    Enforces authentication as a side effect, mirroring existing behavior.
    """
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SessionModel")
    db_pool: Pool = request.app.state.db_pool
    return SessionModel(db_pool)


async def get_slot_model(request: Request) -> SlotModel:
    """Dependency provider returning a `SlotModel` bound to the pool.

    Enforces authentication as a side effect, mirroring existing behavior.
    """
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SlotModel")
    db_pool: Pool = request.app.state.db_pool
    return SlotModel(db_pool)


async def get_shot_model(request: Request) -> ShotModel:
    """Dependency provider returning a `ShotModel` bound to the pool.

    Enforces authentication as a side effect, mirroring existing behavior.
    """
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting ShotModel")
    db_pool: Pool = request.app.state.db_pool
    return ShotModel(db_pool)


async def get_shot_model_ws(websocket: WebSocket) -> ShotModel:
    """Dependency provider returning a `ShotModel` bound to the pool (WebSocket variant)."""
    await require_auth(websocket)
    logger: logging.Logger = websocket.app.state.logger
    logger.debug("Getting ShotModel (WS)")
    db_pool: Pool = websocket.app.state.db_pool
    return ShotModel(db_pool)


async def get_archer_model(request: Request) -> ArcherModel:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting ArcherModel")
    db_pool: Pool = request.app.state.db_pool
    return ArcherModel(db_pool)


async def get_slot_manager(request: Request) -> SlotManager:
    """Dependency provider returning a `SlotManager` for multi-step ops.

    Enforces authentication as a side effect, mirroring existing behavior.
    """
    await require_auth(request)
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SlotManager")
    db_pool: Pool = request.app.state.db_pool
    return SlotManager(db_pool)
