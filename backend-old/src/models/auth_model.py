from datetime import datetime

from asyncpg import Pool

from models.parent_model import ParentModel
from schema import AuthCreate, AuthFilter, AuthRead, AuthSet


class AuthModel(ParentModel[AuthCreate, AuthSet, AuthRead, AuthFilter]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("auth", db_pool, AuthRead)

    async def revoke_by_hash(self, token_hash: bytes, revoked_at: datetime) -> None:
        data = AuthSet(revoked_at=revoked_at)
        where = AuthFilter(session_token_hash=token_hash, revoked_at=None)
        await self.update(data, where=where)
