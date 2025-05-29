from collections.abc import Awaitable, Callable
from typing import Any, Generic, TypeVar

from fastapi import status
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from server.models import DBException, DBNotFound


GenericData = TypeVar("GenericData")  # This will be the inner data type (ArrowsRead, etc.)


class HTTPResponse(BaseModel, Generic[GenericData]):
    code: int
    data: GenericData | None = None
    errors: list[str] = []


async def db_response(
    func: Callable[..., Awaitable[GenericData]],
    success_code: int,
    *args: Any,
) -> JSONResponse:
    try:
        result = await func(*args)
    except DBNotFound as e:
        resp = HTTPResponse[GenericData](
            code=status.HTTP_404_NOT_FOUND,
            data=None,
            errors=[str(e)],
        )
        json_response = JSONResponse(
            status_code=status.HTTP_404_NOT_FOUND,
            content=jsonable_encoder(resp, by_alias=True),
        )
    except DBException as e:
        resp = HTTPResponse[GenericData](
            code=status.HTTP_400_BAD_REQUEST,
            data=None,
            errors=[str(e)],
        )
        json_response = JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content=jsonable_encoder(resp, by_alias=True),
        )
    except Exception as e:
        resp = HTTPResponse[GenericData](
            code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            data=None,
            errors=[str(e)],
        )
        json_response = JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content=jsonable_encoder(resp, by_alias=True),
        )
    else:
        resp = HTTPResponse[GenericData](code=success_code, data=result, errors=[])
        json_response = JSONResponse(
            status_code=success_code,
            content=jsonable_encoder(resp, by_alias=True),
        )
    return json_response
