import uuid
from sqlalchemy import String
from sqlalchemy.dialects.postgresql import UUID, REAL, BOOLEAN
from sqlalchemy.orm import Mapped, mapped_column, relationship

from database.base_model import Base
from database.shots_model import Shots

# pylint: disable=too-few-public-methods


class Arrows(Base):
    __tablename__ = "arrows"

    _id: Mapped[uuid.UUID] = mapped_column(
        "id", UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    human_identifier: Mapped[str] = mapped_column(String(10), nullable=False, unique=True)
    is_programmed: Mapped[bool] = mapped_column(BOOLEAN, nullable=False, default=False)
    weight: Mapped[float | None] = mapped_column(REAL, nullable=True)
    diameter: Mapped[float | None] = mapped_column(REAL, nullable=True)
    spine: Mapped[float | None] = mapped_column(REAL, nullable=True)
    length: Mapped[float | None] = mapped_column(REAL, nullable=True)
    label_position: Mapped[float | None] = mapped_column(REAL, nullable=True)
    shots: Mapped[list[Shots]] = relationship(back_populates="arrow", cascade="all, delete-orphan")
