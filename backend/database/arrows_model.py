import uuid
from sqlalchemy import String
from sqlalchemy.dialects.postgresql import UUID, REAL
from sqlalchemy.orm import Mapped, mapped_column

from database.base_model import Base

# pylint: disable=too-few-public-methods


class Arrows(Base):
    __tablename__ = "arrows"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    human_identifier: Mapped[str] = mapped_column(String(10), nullable=False, unique=True)
    weight: Mapped[float | None] = mapped_column(REAL, nullable=True)
    diameter: Mapped[float | None] = mapped_column(REAL, nullable=True)
    spine: Mapped[float | None] = mapped_column(REAL, nullable=True)
    length: Mapped[float | None] = mapped_column(REAL, nullable=True)
    label_position: Mapped[float | None] = mapped_column(REAL, nullable=True)
