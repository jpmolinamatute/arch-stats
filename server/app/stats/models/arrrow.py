import uuid

from sqlalchemy import Float, ForeignKey, String
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.stats.models import Base


class Arrow(Base):
    __tablename__: str = "arrow"
    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    weight: Mapped[float] = mapped_column(Float, default=0.0)
    diameter: Mapped[float] = mapped_column(Float, default=0.0)
    spine: Mapped[float] = mapped_column(Float, default=0.0)
    length: Mapped[float] = mapped_column(Float, nullable=False)
    human_readable_name: Mapped[str] = mapped_column(String(255), nullable=False)
    archer_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("archer.id", ondelete="CASCADE")
    )
    label_position: Mapped[float] = mapped_column(Float, nullable=False)
