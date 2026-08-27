from __future__ import annotations

from typing import ClassVar

from pydantic import BaseModel, model_validator


class BaseUpdateValidation(BaseModel):
    """
    Base class for all Update schemas.
    Subclasses must set `_id_field_name` to the primary key expected in `where`.
    """

    _id_field_name: ClassVar[str]
    where: BaseModel
    data: BaseModel

    @model_validator(mode="after")
    def _validate_update_payload(self) -> BaseUpdateValidation:
        if len(self.data.model_fields_set) == 0:
            raise ValueError("data must set at least one field")
        if len(self.where.model_fields_set) == 0:
            raise ValueError("where must set at least one field")

        # Ensure the PK is provided in the filter
        if getattr(self.where, self._id_field_name, None) is None:
            raise ValueError(f"where.{self._id_field_name} must be provided")

        return self
