from __future__ import annotations

import uuid

from sqlalchemy import String, CheckConstraint, ForeignKey, UniqueConstraint
from sqlalchemy.dialects.postgresql import UUID, ARRAY, REAL, INTEGER
from sqlalchemy.orm import Mapped, mapped_column, relationship

from database.base_model import Base
from database.sessions_model import Sessions

# pylint: disable=too-few-public-methods


class Targets(Base):
    __tablename__ = "targets"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    max_x_coordinate: Mapped[float] = mapped_column(REAL, nullable=False)
    max_y_coordinate: Mapped[float] = mapped_column(REAL, nullable=False)
    radius: Mapped[list[float]] = mapped_column(ARRAY(REAL), nullable=False)
    points: Mapped[list[int]] = mapped_column(ARRAY(INTEGER), nullable=False)
    height: Mapped[float] = mapped_column(REAL, nullable=False)
    human_identifier: Mapped[str | None] = mapped_column(String(10))
    session_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("sessions.id", ondelete="CASCADE"), nullable=False
    )

    __table_args__ = (
        CheckConstraint(
            "array_length(radius, 1) = array_length(points, 1)", name="radius_points_length_check"
        ),
        UniqueConstraint("session_id", "human_identifier", name="unique_session_human_identifier"),
    )

    session: Mapped[Sessions] = relationship(back_populates="targets")
