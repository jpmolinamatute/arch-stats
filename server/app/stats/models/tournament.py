import uuid

from sqlalchemy import TIMESTAMP, Boolean, Integer
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.stats.models import Base


class Tournament(Base):
    __tablename__: str = "tournament"
    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    start_time: Mapped[str] = mapped_column(TIMESTAMP, nullable=False)
    end_time: Mapped[str] = mapped_column(TIMESTAMP)
    number_of_lanes: Mapped[int] = mapped_column(Integer, nullable=False)
    is_opened: Mapped[bool] = mapped_column(Boolean, default=False)
