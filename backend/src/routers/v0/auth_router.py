import base64
import logging
from datetime import UTC, datetime
from uuid import UUID

from asyncpg import Pool
from fastapi import APIRouter, Depends, HTTPException, Request, Response, status
from jwt import ExpiredSignatureError, InvalidTokenError

from core import (
    AuthDeps,
    build_needs_registration_response,
    decode_token,
    hash_session_token,
    login_existing_archer,
    register_archer,
    settings,
    verify_google_id_token,
)
from models import ArcherModel, AuthModel, DBNotFound
from schema import (
    ArcherFilter,
    AuthAuthenticated,
    AuthNeedsRegistration,
    AuthRegistrationRequest,
    AuthStatus,
    GoogleOneTapRequest,
    LogoutResponse,
)


router = APIRouter(prefix="/auth", tags=["Auth"])


async def get_archer_model(request: Request) -> ArcherModel:
    """Dependency provider for `ArcherModel`.

    Args:
        request: FastAPI request object (for app state access).

    Returns:
        ArcherModel: Model bound to the shared asyncpg pool.
    """
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting ArcherModel")
    db_pool: Pool = request.app.state.db_pool
    return ArcherModel(db_pool)


async def get_auth_model(request: Request) -> AuthModel:
    """Dependency provider for `AuthModel`.

    Args:
        request: FastAPI request.

    Returns:
        AuthModel bound to the shared asyncpg pool.
    """
    logger: logging.Logger = request.app.state.logger
    logger.debug("Getting AuthModel")
    db_pool: Pool = request.app.state.db_pool
    return AuthModel(db_pool)


async def get_deps(request: Request) -> AuthDeps:
    archers = await get_archer_model(request)
    sessions = await get_auth_model(request)
    return AuthDeps(archers=archers, sessions=sessions)


@router.post(
    "/google",
    # Support both 201 (authenticated) and 200 (needs registration) shapes at runtime
    response_model=AuthNeedsRegistration | AuthAuthenticated,
    status_code=status.HTTP_201_CREATED,
    responses={
        200: {
            "model": AuthNeedsRegistration,
            "description": "Archer was authenticated but needs registration",
        },
        201: {
            "model": AuthAuthenticated,
            "description": "Archer authenticated successfully",
        },
    },
)
@router.post(
    "/login",
    # Support both 201 (authenticated) and 200 (needs registration) shapes at runtime
    response_model=AuthNeedsRegistration | AuthAuthenticated,
    status_code=status.HTTP_201_CREATED,
    responses={
        200: {
            "model": AuthNeedsRegistration,
            "description": "Archer was authenticated but needs registration",
        },
        201: {
            "model": AuthAuthenticated,
            "description": "Archer authenticated successfully",
        },
    },
)
async def google_one_tap_login(
    payload: GoogleOneTapRequest,
    response: Response,
    request: Request,
    archer_model: ArcherModel = Depends(get_archer_model),
    auth_session_model: AuthModel = Depends(get_auth_model),
) -> AuthNeedsRegistration | AuthAuthenticated:
    """Verify Google One Tap; start a session if user exists.

    - Existing archer: create session, set `arch_stats_auth` HttpOnly cookie, return
      AuthAuthenticated (201).
    - Unknown subject: return AuthNeedsRegistration (200) with prefilled Google data.

    Responses: 201 (authenticated), 200 (needs registration), 401 (invalid token), 400
    (missing claims).
    """
    try:
        logger: logging.Logger = request.app.state.logger
        logger.info("Initializing Google One Tap credential process")
        user_info = await verify_google_id_token(payload.credential)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail=str(e)) from e

    bad_request = []
    sub: str | None = user_info.get("sub", None)
    email: str | None = user_info.get("email", None)
    if not isinstance(sub, str) and not sub:
        bad_request.append("subject_missing")

    if not isinstance(email, str) and not email:
        bad_request.append("email_missing")

    if bad_request:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=bad_request)

    now = datetime.now(UTC)
    assert isinstance(sub, str)  # this is just to convince mypy
    assert isinstance(email, str)  # this is just to convince mypy
    archer = await archer_model.get_by_google_subject(sub)
    if archer is None:
        response.status_code = status.HTTP_200_OK
        return build_needs_registration_response(user_info)

    # Archer exists -> create session + cookie and return authenticated
    authd = await login_existing_archer(
        auth_deps=AuthDeps(archers=archer_model, sessions=auth_session_model),
        user_info=user_info,
        archer=archer,
        now=now,
        user_agent=request.headers.get("user-agent"),
        ip=(request.client.host if request.client else None),
    )
    # Set cookie here (service returns token/expiry only)
    cookie_secure = request.url.scheme == "https"
    response.set_cookie(
        key="arch_stats_auth",
        value=authd.access_token,
        httponly=True,
        secure=cookie_secure,
        samesite="lax",
        path="/",
        max_age=settings.jwt_ttl_minutes * 60,
    )
    response.status_code = status.HTTP_201_CREATED
    return authd


@router.get(
    "/me",
    response_model=AuthAuthenticated,
    status_code=status.HTTP_200_OK,
)
async def get_current_user(
    request: Request,
    archer_model: ArcherModel = Depends(get_archer_model),
) -> AuthAuthenticated:
    """Get the currently authenticated archer from the JWT cookie.

    Reads the `arch_stats_auth` cookie, decodes the JWT to extract the archer_id,
    fetches the archer details, and returns an AuthAuthenticated response.

    Responses: 200 OK (authenticated), 401 (no cookie or invalid token).
    """
    token = request.cookies.get("arch_stats_auth")
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="No authentication cookie found"
        )

    try:
        archer_id_str = decode_token(token, "sub")
        if not isinstance(archer_id_str, str):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token: missing subject"
            )

        # Extract expiration timestamp from JWT
        exp_timestamp = decode_token(token, "exp")
        if not isinstance(exp_timestamp, (int, float)):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token: missing expiration"
            )

        # Convert Unix timestamp to datetime
        expires_at = datetime.fromtimestamp(exp_timestamp, tz=UTC)

        archer_id = UUID(archer_id_str)

        try:
            where = ArcherFilter(archer_id=archer_id)
            archer = await archer_model.get_one(where)
        except DBNotFound as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Archer not found"
            ) from e

        # Return the same shape as login endpoint
        return AuthAuthenticated(
            archer=archer,
            access_token=token,  # Return the same token
            status=AuthStatus.AUTHENTICATED,
            expires_at=expires_at,
        )
    except ExpiredSignatureError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Token has expired"
        ) from e
    except InvalidTokenError as e:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token") from e
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail=f"Authentication failed: {str(e)}"
        ) from e


@router.post(
    "/logout",
    response_model=LogoutResponse,
)
async def logout(
    response: Response,
    request: Request,
    auth_session_model: AuthModel = Depends(get_auth_model),
) -> LogoutResponse:
    """Log out: clear cookie and revoke session if present.

    - Deletes the `arch_stats_auth` HttpOnly cookie.
    - If a valid JWT is supplied, revokes the matching server-side session; otherwise ignored.
    - Idempotent: always returns `LogoutResponse(success=True)`.

    Responses: 200 OK.
    """
    logger: logging.Logger = request.app.state.logger
    logger.info("Login Archer out")
    token = request.cookies.get("arch_stats_auth")
    if token:
        try:
            sid_b64 = decode_token(token, "sid")
            if isinstance(sid_b64, str):
                raw = base64.urlsafe_b64decode(sid_b64 + "==")
                token_hash = hash_session_token(raw)
                await auth_session_model.revoke_by_hash(token_hash, datetime.now(UTC))
        except Exception:
            pass
    response.delete_cookie("arch_stats_auth", path="/")
    response.status_code = status.HTTP_200_OK
    return LogoutResponse(success=True)


@router.post(
    "/register",
    response_model=AuthAuthenticated,
    status_code=status.HTTP_201_CREATED,
    responses={
        200: {"model": AuthAuthenticated, "description": "User was already registered"},
    },
)
async def register(
    payload: AuthRegistrationRequest,
    response: Response,
    request: Request,
    auth_deps: AuthDeps = Depends(get_deps),
) -> AuthAuthenticated:
    """Register a new archer (or log in if already exists).

    - Verifies Google credential. If an archer with the subject exists, authenticate and set
      `arch_stats_auth` cookie (200).
    - Otherwise, validate names, create archer, authenticate, and set cookie (201).

    Responses: 201 (created + authenticated), 200 (authenticated existing),
    401 (invalid token), 400 (invalid Google info or missing names).
    """

    try:
        user_info = await verify_google_id_token(payload.credential)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail=str(e)) from e

    sub = user_info.get("sub")
    email: str | None = user_info.get("email")
    if not isinstance(sub, str) or not email:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid Google info")

    existing = await auth_deps.archers.get_by_google_subject(sub)
    now = datetime.now(UTC)
    user_agent = request.headers.get("user-agent")
    ip = request.client.host if request.client else None

    if existing is None:
        try:
            result = await register_archer(
                auth_deps=auth_deps,
                payload=payload,
                user_info=user_info,
                now=now,
                user_agent=user_agent,
                ip=ip,
            )
        except ValueError as e:
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e)) from e
    else:
        result = await login_existing_archer(
            auth_deps=auth_deps,
            user_info=user_info,
            archer=existing,
            now=now,
            user_agent=user_agent,
            ip=ip,
        )
    data = result.model_dump(exclude_unset=True, exclude_none=True)
    response.set_cookie(
        key="arch_stats_auth",
        value=data["access_token"],
        httponly=True,
        secure=request.url.scheme == "https",
        samesite="lax",
        path="/",
        max_age=settings.jwt_ttl_minutes * 60,
    )
    response.status_code = status.HTTP_200_OK if existing else status.HTTP_201_CREATED
    return result
