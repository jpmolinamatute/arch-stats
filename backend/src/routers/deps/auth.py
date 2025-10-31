from uuid import UUID

from fastapi import HTTPException, Request, status

from core import decode_token


async def require_auth(request: Request) -> UUID:
    """Validate JWT from auth cookie and return authenticated archer id.

    Raises 401 if the cookie is missing or invalid.
    """
    token = request.cookies.get("arch_stats_auth")
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User is not authorized to use this endpoint",
        )
    try:
        sub = decode_token(token, "sub")
        if not isinstance(sub, str):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="User is not authorized to use this endpoint",
            )
        return UUID(sub)
    except Exception as exc:  # pylint: disable=broad-except
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User is not authorized to use this endpoint",
        ) from exc
