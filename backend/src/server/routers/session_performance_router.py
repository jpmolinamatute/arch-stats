import logging

from asyncpg import Pool
from fastapi import APIRouter, Depends, Request, status
from fastapi.responses import JSONResponse

from server.routers.utils import HTTPResponse, db_response
from shared.models import SessionPerformanceModel
from shared.schema import SessionPerformanceRead


SessionPerformanceRouter = APIRouter(prefix="/session_performance")


async def get_session_performance_db(request: Request) -> SessionPerformanceModel:
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting SessionPerformanceModel")
    db_pool: Pool = request.app.state.db_pool
    return SessionPerformanceModel(db_pool)


@SessionPerformanceRouter.get("", response_model=HTTPResponse[list[SessionPerformanceRead]])
async def get_session_performance(
    sp_db: SessionPerformanceModel = Depends(get_session_performance_db),
) -> JSONResponse:
    """
    Retrieve performance metrics for the currently open session.

    Returns a list of enriched shot records (scoring, speed, flight time, etc.).
    """
    return await db_response(sp_db.get_all, status.HTTP_200_OK)
