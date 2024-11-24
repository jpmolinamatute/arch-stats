import uuid

from sqlalchemy import TIMESTAMP, Enum, Float, String
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.stats.models import Base


class Archer(Base):
    __tablename__: str = "archer"
    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    first_name: Mapped[str] = mapped_column(String(255), nullable=False)
    last_name: Mapped[str] = mapped_column(String(255), nullable=False)
    password: Mapped[str] = mapped_column(String(255), nullable=False)
    email: Mapped[str] = mapped_column(String(255), unique=True, nullable=False)
    genre: Mapped[str] = mapped_column(
        Enum("male", "female", "no-answered", name="archer_genre"), nullable=False
    )
    type_of_archer: Mapped[str] = mapped_column(
        Enum("compound", "traditional", "barebow", "olympic", name="archer_type"), nullable=False
    )
    bow_poundage: Mapped[float] = mapped_column(Float, nullable=False)
    created_at: Mapped[str] = mapped_column(TIMESTAMP, default="current_timestamp")
