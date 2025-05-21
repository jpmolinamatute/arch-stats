from typing import Generic, TypeVar

from pydantic import BaseModel


GenericData = TypeVar("GenericData")  # This will be the inner data type (ArrowRead, etc.)


class HTTPResponse(BaseModel, Generic[GenericData]):
    code: int
    data: GenericData | None = None
    errors: list[str | None] = []
