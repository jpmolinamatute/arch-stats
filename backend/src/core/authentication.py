"""Authentication core utilities and services.

This module centralizes auth-related logic so FastAPI routers remain thin and
focused on HTTP concerns (dependency injection, cookie handling, response
envelopes). All functions here are pure or DB-bound and do not manipulate HTTP
objects directly.

Rules:
- Keep functions fully typed and framework-agnostic.
- Accept primitives (strings, datetimes) instead of Request/Response.
- Return domain models or simple primitives that routers can translate into
  HTTP responses and cookies.
"""

import asyncio
import base64
import hashlib
import os
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import NotRequired, TypedDict, cast
from uuid import UUID

import jwt
from google.auth.transport import requests as google_auth_requests
from google.oauth2 import id_token

from core.settings import settings
from models import ArcherModel, AuthModel
from schema import (
    ArcherCreate,
    ArcherFilter,
    ArcherRead,
    ArcherSet,
    AuthAuthenticated,
    AuthCreate,
    AuthNeedsRegistration,
    AuthRegistrationRequest,
)


# Convenience structure to pass around models together
@dataclass(frozen=True)
class AuthDeps:
    archers: ArcherModel
    sessions: AuthModel


class GoogleUserData(TypedDict):
    """Subset of Google ID Token (OIDC) claims used by the app."""

    sub: str
    email: str
    email_verified: NotRequired[bool]
    name: NotRequired[str]
    given_name: NotRequired[str]
    family_name: NotRequired[str]
    picture: NotRequired[str]
    locale: NotRequired[str]
    hd: NotRequired[str]

    aud: NotRequired[str | list[str]]
    azp: NotRequired[str]
    iss: NotRequired[str]
    iat: NotRequired[int]
    exp: NotRequired[int]
    nonce: NotRequired[str]


def hash_session_token(raw: bytes) -> bytes:
    """Return SHA-256 digest of the raw session token bytes.

    Args:
        raw: Random session token bytes.

    Returns:
        32-byte SHA-256 digest.
    """

    return hashlib.sha256(raw).digest()


def decode_token(token: str, attr_name: str) -> str | int | float | None:
    """Decode a JWT token and extract a specific attribute.

    Args:
        token: The JWT token string to decode.
        attr_name: The attribute name to extract from the decoded token (e.g., "sub", "sid", "exp").

    Returns:
        The value of the requested attribute, or None if not present.

    Raises:
        jwt.InvalidTokenError: If the token is invalid or verification fails.
        jwt.ExpiredSignatureError: If the token has expired.
    """
    decoded = jwt.decode(
        token, settings.arch_stats_jwt_secret, algorithms=[settings.arch_stats_jwt_algorithm]
    )
    value: str | int | float | None = decoded.get(attr_name, None)
    return value


async def verify_google_id_token(credential: str) -> GoogleUserData:
    """Verify a Google One Tap credential and return its ID token claims.

    The verification performs network I/O; run it in a worker thread to avoid
    blocking the event loop.

    Raises ValueError when verification fails.
    """
    try:
        req = google_auth_requests.Request()  # type: ignore[no-untyped-call]
        verified = await asyncio.to_thread(
            id_token.verify_oauth2_token,
            credential,
            req,
            settings.arch_stats_google_oauth_client_id,
        )
        return cast(GoogleUserData, dict(verified))
    except Exception as exc:
        # Normalize all verification failures
        # The google-auth library may raise a variety of exceptions depending on
        # network conditions or token shape. Normalize to ValueError so routers
        # consistently return 401 instead of 500.
        raise ValueError(f"Invalid Google credential: {exc}") from exc


def build_jwt(archer_id: UUID, sid_b64: str, issued_at: datetime, expires_at: datetime) -> str:
    """Build and sign the access JWT embedding archer id and session id."""

    payload = {
        "sub": str(archer_id),
        "sid": sid_b64,
        "exp": int(expires_at.timestamp()),
        "iat": int(issued_at.timestamp()),
        "iss": "arch-stats",
        "typ": "access",
    }
    return jwt.encode(
        payload, settings.arch_stats_jwt_secret, algorithm=settings.arch_stats_jwt_algorithm
    )


async def create_auth_session(
    auth: AuthModel,
    archer_id: UUID,
    now: datetime,
    *,
    user_agent: str | None,
    ip: str | None,
) -> tuple[bytes, datetime]:
    """Insert auth session row and return (raw_session, expires_at)."""

    raw_session = os.urandom(settings.session_token_bytes)
    token_hash = hash_session_token(raw_session)
    expires_at = now + timedelta(minutes=settings.arch_stats_jwt_ttl_minutes)

    data = AuthCreate(
        archer_id=archer_id,
        session_token_hash=token_hash,
        created_at=now,
        expires_at=expires_at,
        ua=user_agent,
        ip_inet=ip,
    )
    await auth.insert_one(data)
    return raw_session, expires_at


async def update_archer_last_login(
    archers: ArcherModel, user_info: GoogleUserData, now: datetime, archer_id: UUID
) -> None:
    """Update last login and picture URL from Google claims."""

    picture = user_info.get("picture") or ""
    data = ArcherSet(
        google_picture_url=picture,
        last_login_at=now,
    )
    where = ArcherFilter(archer_id=archer_id)
    await archers.update(data, where)


def build_needs_registration_response(user_info: GoogleUserData) -> AuthNeedsRegistration:
    """Compose the AuthNeedsRegistration payload from Google claims."""

    given = user_info.get("given_name", None)
    family = user_info.get("family_name", None)
    picture = user_info.get("picture", None)
    return AuthNeedsRegistration(
        google_email=user_info["email"],
        google_subject=user_info["sub"],
        given_name=given,
        family_name=family,
        given_name_provided=bool(given and given.strip()),
        family_name_provided=bool(family and family.strip()),
        picture_url=picture,
    )


async def authenticate_archer(
    *,
    auth_deps: AuthDeps,
    archer: ArcherRead,
    now: datetime,
    user_agent: str | None,
    ip: str | None,
) -> AuthAuthenticated:
    """Create a session, build JWT and return an AuthAuthenticated payload.

    Note: no HTTP cookie is set here; routers handle cookie headers.
    """

    raw_session, expires_at = await create_auth_session(
        auth_deps.sessions, archer.get_id(), now, user_agent=user_agent, ip=ip
    )
    sid_b64 = base64.urlsafe_b64encode(raw_session).decode().rstrip("=")
    jwt_token = build_jwt(archer.get_id(), sid_b64, now, expires_at)
    return AuthAuthenticated(access_token=jwt_token, expires_at=expires_at, archer=archer)


async def login_existing_archer(
    *,
    auth_deps: AuthDeps,
    user_info: GoogleUserData,
    archer: ArcherRead,
    now: datetime,
    user_agent: str | None,
    ip: str | None,
) -> AuthAuthenticated:
    """Update last login for an existing archer and authenticate them."""

    await update_archer_last_login(auth_deps.archers, user_info, now, archer.get_id())
    # Re-fetch to ensure we return latest state (e.g., picture/last_login updated)
    where = ArcherFilter(archer_id=archer.get_id())
    archer = await auth_deps.archers.get_one(where)
    return await authenticate_archer(
        auth_deps=auth_deps, archer=archer, now=now, user_agent=user_agent, ip=ip
    )


async def register_archer(
    *,
    auth_deps: AuthDeps,
    payload: AuthRegistrationRequest,
    user_info: GoogleUserData,
    now: datetime,
    user_agent: str | None,
    ip: str | None,
) -> AuthAuthenticated:
    """Register a new archer and return the auth payload.

    Performs minimal validation on names, creates the Archer row, and
    authenticates the new archer. Framework-agnostic: raises `ValueError` for
    validation errors; routers should translate to HTTP errors.
    """

    given = (payload.first_name or user_info.get("given_name") or "").strip()
    family = (payload.last_name or user_info.get("family_name") or "").strip()
    missing_names: list[str] = []
    if not given:
        missing_names.append("given_name is missing")
    if not family:
        missing_names.append("family_name is missing")

    if missing_names:
        raise ValueError(", ".join(missing_names))

    archer_id = await auth_deps.archers.insert_one(
        ArcherCreate(
            email=user_info["email"],
            first_name=given,
            last_name=family,
            date_of_birth=payload.date_of_birth,
            gender=payload.gender,
            bowstyle=payload.bowstyle,
            draw_weight=payload.draw_weight,
            google_picture_url=(user_info.get("picture") or ""),
            google_subject=user_info["sub"],
        )
    )

    # Authenticate newly created archer
    where = ArcherFilter(archer_id=archer_id)
    archer = await auth_deps.archers.get_one(where)
    return await authenticate_archer(
        auth_deps=auth_deps,
        archer=archer,
        now=now,
        user_agent=user_agent,
        ip=ip,
    )
