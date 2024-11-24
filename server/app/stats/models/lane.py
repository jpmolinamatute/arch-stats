import uuid

from sqlalchemy import Float, ForeignKey, Integer
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.stats.models import Base


class Lane(Base):
    __tablename__: str = "lane"
    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    tournament_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("tournaments.id", ondelete="CASCADE")
    )
    lane_number: Mapped[int] = mapped_column(Integer, nullable=False)
    max_x_coordinate: Mapped[float] = mapped_column(Float)
    max_y_coordinate: Mapped[float] = mapped_column(Float)
    number_of_archers: Mapped[int] = mapped_column(Integer, nullable=False)
