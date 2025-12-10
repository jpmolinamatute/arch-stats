import datetime
from http import HTTPStatus
from unittest.mock import AsyncMock, patch
from uuid import uuid4

import pytest
from httpx import AsyncClient

from core import settings
from schema import ArcherRead, BowStyleType, GenderType


@pytest.mark.asyncio
async def test_dummy_login_creates_new_archer(
    client: AsyncClient,
    setup_auth_deps: None,  # pylint: disable=unused-argument
    mock_archers: AsyncMock,
    mock_sessions: AsyncMock,
) -> None:
    """Test dummy login creates a new archer when one doesn't exist."""

    # Setup mocks
    # 1. get_by_google_subject returns None (not found)
    mock_archers.get_by_google_subject.return_value = None

    # 2. insert_one returns a new UUID
    new_id = uuid4()
    mock_archers.insert_one.return_value = new_id

    # 3. get_one returns the created archer (needed for login)
    dummy_archer = ArcherRead(
        archer_id=new_id,
        email="dummy@example.com",
        first_name="Dummy",
        last_name="Archer",
        created_at=datetime.datetime.now(),
        last_login_at=datetime.datetime.now(),
        date_of_birth=datetime.date(2000, 1, 1),
        gender=GenderType.UNSPECIFIED,
        bowstyle=BowStyleType.RECURVE,
        draw_weight=30.0,
        google_picture_url="",
        google_subject="dummy_subject",
        # Add other required fields if any, checking schema
    )
    mock_archers.get_one.return_value = dummy_archer

    # 4. update (last_login)
    mock_archers.update.return_value = None

    # 5. sessions.insert_one
    mock_sessions.insert_one.return_value = None

    with patch.object(settings, "arch_stats_dev_mode", True):
        response = await client.post("/api/v0/auth/dummy")

        # Validation
        assert response.status_code == HTTPStatus.CREATED
        data = response.json()
        assert data["archer"]["email"] == "dummy@example.com"
        assert data["archer"]["first_name"] == "Dummy"
        assert "access_token" in data

        # Verify calls
        mock_archers.get_by_google_subject.assert_called_with("dummy_subject")
        mock_archers.insert_one.assert_called_once()
        # Verify created data has correct hardcoded values
        call_args = mock_archers.insert_one.call_args[0][0]
        assert call_args.email == "dummy@example.com"
        assert call_args.first_name == "Dummy"

        # Verify cookie is set correctly
        assert "arch_stats_auth" in response.cookies
        cookie = next(c for c in response.cookies.jar if c.name == "arch_stats_auth")
        assert cookie.secure is False  # Localhost/http
        assert cookie.path == "/"


@pytest.mark.asyncio
async def test_dummy_login_existing_archer(
    client: AsyncClient,
    setup_auth_deps: None,  # pylint: disable=unused-argument
    mock_archers: AsyncMock,
    mock_sessions: AsyncMock,
) -> None:
    """Test dummy login logs in existing archer if already created."""

    # Setup mocks
    existing_id = uuid4()
    existing_archer = ArcherRead(
        archer_id=existing_id,
        email="dummy@example.com",
        first_name="Dummy",
        last_name="Archer",
        created_at=datetime.datetime.now(),
        last_login_at=datetime.datetime.now(),
        date_of_birth=datetime.date(2000, 1, 1),
        gender=GenderType.UNSPECIFIED,
        bowstyle=BowStyleType.RECURVE,
        draw_weight=30.0,
        google_picture_url="",
        google_subject="dummy_subject",
    )

    # 1. get_by_google_subject returns existing archer
    mock_archers.get_by_google_subject.return_value = existing_archer

    # 2. get_one (called during login re-fetch)
    mock_archers.get_one.return_value = existing_archer

    # 3. update and session insert
    mock_archers.update.return_value = None
    mock_sessions.insert_one.return_value = None

    with patch.object(settings, "arch_stats_dev_mode", True):
        response = await client.post("/api/v0/auth/dummy")

        assert response.status_code == HTTPStatus.CREATED
        data = response.json()
        assert data["archer"]["archer_id"] == str(existing_id)

        # Verify insert_one was NOT called
        mock_archers.insert_one.assert_not_called()
        # Verify update was called (last_login)
        mock_archers.update.assert_called_once()


@pytest.mark.asyncio
async def test_dummy_login_disabled_in_prod(
    client: AsyncClient,
    setup_auth_deps: None,  # pylint: disable=unused-argument
) -> None:
    """Test dummy login fails when dev mode is False."""
    with patch.object(settings, "arch_stats_dev_mode", False):
        response = await client.post("/api/v0/auth/dummy")
        assert response.status_code == HTTPStatus.NOT_FOUND
