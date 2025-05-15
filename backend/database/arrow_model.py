import uuid
from sqlalchemy import String, Float
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from database.base_model import Base

# pylint: disable=too-few-public-methods


class Arrow(Base):
    __tablename__ = "arrows"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    human_identifier: Mapped[str] = mapped_column(String(10), nullable=False, unique=True)
    weight: Mapped[float] = mapped_column(Float, nullable=True)
    diameter: Mapped[float] = mapped_column(Float, nullable=True)
    spine: Mapped[float] = mapped_column(Float, nullable=True)
    length: Mapped[float] = mapped_column(Float, nullable=True)
    label_position: Mapped[float] = mapped_column(Float, nullable=True)
