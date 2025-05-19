from __future__ import annotations

import uuid
from datetime import datetime

from sqlalchemy import String, CheckConstraint
from sqlalchemy.dialects.postgresql import UUID, TIMESTAMP, BOOLEAN
from sqlalchemy.orm import Mapped, mapped_column, relationship

from database.base_model import Base
from database.targets_model import Targets


# pylint: disable=too-few-public-methods
class Sessions(Base):
    __tablename__ = "sessions"

    _id: Mapped[uuid.UUID] = mapped_column(
        "id", UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    is_opened: Mapped[bool] = mapped_column(BOOLEAN, nullable=False)
    start_time: Mapped[datetime] = mapped_column(TIMESTAMP(timezone=True), nullable=False)
    end_time: Mapped[datetime | None] = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    location: Mapped[str] = mapped_column(String(255), nullable=False)

    __table_args__ = (
        CheckConstraint(
            "(is_opened = TRUE AND end_time IS NULL) OR (is_opened = FALSE)",
            name="open_session_no_end_time",
        ),
        CheckConstraint(
            "(is_opened = FALSE AND end_time IS NOT NULL) OR (is_opened = TRUE)",
            name="closed_session_with_end_time",
        ),
    )

    targets: Mapped[list[Targets]] = relationship(
        back_populates="session", cascade="all, delete-orphan"
    )
