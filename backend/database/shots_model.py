from __future__ import annotations

import uuid
from datetime import datetime


from sqlalchemy import ForeignKey
from sqlalchemy.dialects.postgresql import UUID, TIMESTAMP, REAL
from sqlalchemy.orm import Mapped, mapped_column, relationship

from database.base_model import Base
from database.arrows_model import Arrows
from database.sessions_model import Sessions


# pylint: disable=too-few-public-methods
class Shots(Base):
    __tablename__ = "shots"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)

    arrow_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("arrows.id", ondelete="CASCADE"), nullable=False
    )

    session_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("sessions.id", ondelete="CASCADE"), nullable=False
    )

    arrow_engage_time: Mapped[datetime] = mapped_column(TIMESTAMP(timezone=True), nullable=False)
    arrow_disengage_time: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True),
        nullable=False,
    )
    arrow_landing_time: Mapped[datetime | None] = mapped_column(
        TIMESTAMP(timezone=True), nullable=True
    )

    x_coordinate: Mapped[float | None] = mapped_column(REAL, nullable=True)
    y_coordinate: Mapped[float | None] = mapped_column(REAL, nullable=True)

    # Relationships (optional, if needed)
    arrow: Mapped[Arrows] = relationship()
    session: Mapped[Sessions] = relationship()
