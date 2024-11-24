import uuid

from sqlalchemy import BigInteger, Float, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.stats.models import Base


class Shot(Base):
    __tablename__: str = "shot"
    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    arrow_engage_time: Mapped[int] = mapped_column(BigInteger)
    arrow_disengage_time: Mapped[int] = mapped_column(BigInteger)
    arrow_landing_time: Mapped[int] = mapped_column(BigInteger)
    x_coordinate: Mapped[float] = mapped_column(Float)
    y_coordinate: Mapped[float] = mapped_column(Float)
    pull_length: Mapped[float] = mapped_column(Float)
    distance: Mapped[float] = mapped_column(Float)
    lane_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("lane.id", ondelete="CASCADE")
    )
    arrow_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("arrow.id", ondelete="CASCADE")
    )
