from typing import Generic, Any, TypeVar
from collections.abc import Callable, Awaitable
from fastapi import status

from pydantic import BaseModel

from database import DBNotFound, DBException

GenericData = TypeVar("GenericData")  # This will be the inner data type (ArrowsRead, etc.)


class HTTPResponse(BaseModel, Generic[GenericData]):
    code: int
    data: GenericData | None = None
    errors: list[str | None] = []


async def db_response(
    func: Callable[..., Awaitable[GenericData]],
    success_code: int,
    *args: Any,
) -> HTTPResponse[GenericData]:
    try:
        result = await func(*args)
        return HTTPResponse[GenericData](code=success_code, data=result, errors=[])
    except DBNotFound as e:
        return HTTPResponse[GenericData](code=status.HTTP_404_NOT_FOUND, data=None, errors=[str(e)])
    except DBException as e:
        return HTTPResponse[GenericData](
            code=status.HTTP_400_BAD_REQUEST, data=None, errors=[str(e)]
        )
    except Exception as e:
        return HTTPResponse[GenericData](
            code=status.HTTP_500_INTERNAL_SERVER_ERROR, data=None, errors=[str(e)]
        )
