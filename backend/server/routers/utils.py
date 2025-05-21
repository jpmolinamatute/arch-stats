from typing import Any, TypeVar
from collections.abc import Callable, Awaitable
from fastapi import status

from database.base import DBNotFound, DBException
from database.schema import HTTPResponse

T = TypeVar("T")


async def db_response(
    func: Callable[..., Awaitable[T]],
    success_code: int,
    *args: Any,
) -> HTTPResponse[T]:
    try:
        result = await func(*args)
        return HTTPResponse[T](code=success_code, data=result, errors=[])
    except DBNotFound as e:
        return HTTPResponse[T](code=status.HTTP_404_NOT_FOUND, data=None, errors=[str(e)])
    except DBException as e:
        return HTTPResponse[T](code=status.HTTP_400_BAD_REQUEST, data=None, errors=[str(e)])
    except Exception as e:
        return HTTPResponse[T](
            code=status.HTTP_500_INTERNAL_SERVER_ERROR, data=None, errors=[str(e)]
        )
