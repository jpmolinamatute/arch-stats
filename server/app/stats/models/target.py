import uuid

from sqlalchemy import ARRAY, Float, ForeignKey, Integer, String
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.stats.models import Base


class Target(Base):
    __tablename__: str = "target"
    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    x_coordinate: Mapped[float] = mapped_column(Float)
    y_coordinate: Mapped[float] = mapped_column(Float)
    radius: Mapped[list[float]] = mapped_column(ARRAY(Float))
    points: Mapped[list[int]] = mapped_column(ARRAY(Integer))
    height: Mapped[float] = mapped_column(Float)
    human_readable_name: Mapped[str] = mapped_column(String(255))
    lane_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("lane.id", ondelete="CASCADE")
    )
    archer_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("archers.id", ondelete="CASCADE")
    )
