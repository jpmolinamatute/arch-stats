from uuid import UUID

from fastapi import HTTPException, status

from core.base_manager import BaseManager
from models.parent_model import DBException, DBNotFound
from schema import SessionCreate, SessionFilter, SessionId, SessionRead


class SessionManager(BaseManager):
    """Business logic for sessions."""

    async def get_open_session_for_archer(
        self, current_archer_id: UUID, archer_id: UUID
    ) -> SessionId:
        self.verify_archer_identity(current_archer_id, archer_id)
        session_id = await self.session.archer_open_session(archer_id)
        return SessionId(session_id=session_id)

    async def get_closed_session_for_archer(
        self, current_archer_id: UUID, archer_id: UUID
    ) -> list[SessionRead]:
        self.verify_archer_identity(current_archer_id, archer_id)
        return await self.session.get_all_closed_sessions_owned_by(archer_id)

    async def get_participating_session_for_archer(
        self, current_archer_id: UUID, archer_id: UUID
    ) -> SessionId:
        self.verify_archer_identity(current_archer_id, archer_id)
        session_id = await self.session.is_archer_participating(archer_id)
        return SessionId(session_id=session_id)

    async def get_all_open_sessions(self) -> list[SessionRead]:
        return await self.session.get_all_open_sessions()

    async def create_session(
        self, session_data: SessionCreate, current_archer_id: UUID
    ) -> SessionId:
        # Identity check: only the authenticated archer can open a session for themselves
        if current_archer_id != session_data.owner_archer_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="ERROR: user not allowed to open a session for another archer",
            )
        try:
            session_id = await self.session.create_session(session_data)
            return SessionId(session_id=session_id)
        except ValueError as e:
            msg = str(e)
            if msg == "Archer already has an opened session":
                raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=msg) from e
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
            ) from e

    async def get_session(self, session_id: UUID, current_archer_id: UUID) -> SessionRead:
        try:
            session = await self.session.get_one(SessionFilter(session_id=session_id))
            if not session.is_opened and session.owner_archer_id != current_archer_id:
                raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")
            return session
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e

    async def re_open_session(self, session: SessionId, current_archer_id: UUID) -> SessionId:
        try:
            await self.session.re_open_session(session, current_archer_id)
            return session
        except DBNotFound as e:
            # Standardize to 422 for non-existent or already-open sessions
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
                detail="ERROR: Session either doesn't exist or it was already open",
            ) from e
        except DBException as e:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
        except ValueError as e:
            msg = str(e)
            if msg == "Archer already has an opened session":
                raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=msg) from e
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=str(e)
            ) from e

    async def close_session(self, session: SessionId, current_archer_id: UUID) -> dict[str, str]:
        try:
            await self.session.close_session(session, current_archer_id)
            return {"status": "closed"}
        except DBNotFound as e:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e)) from e
        except DBException as e:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e)) from e
        except ValueError as e:
            msg = str(e)
            if msg == "ERROR: session_id wasn't provided":
                raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=msg) from e
            if msg == "ERROR: cannot close session with active participants":
                raise HTTPException(
                    status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
                ) from e
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_CONTENT, detail=msg
            ) from e
