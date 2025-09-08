import logging
from abc import ABC
from datetime import datetime
from typing import Generic, Protocol, TypeVar
from uuid import UUID

from asyncpg import Pool
from pydantic import BaseModel

from shared.logger import LogLevel, get_logger
from shared.settings import settings


Values = str | float | bool | int | UUID | datetime | None | list[int] | list[float]
ValuesList = list[Values]


class HasId(Protocol):
    def get_id(self) -> UUID: ...


CREATETYPE = TypeVar("CREATETYPE", bound=BaseModel)
UPDATETYPE = TypeVar("UPDATETYPE", bound=BaseModel)
READTYPE = TypeVar("READTYPE", bound=HasId)
FILTERTYPE = TypeVar("FILTERTYPE", bound=BaseModel)


class DBException(Exception):
    pass


class DBNotFound(Exception):
    pass


class ParentModel(Generic[CREATETYPE, UPDATETYPE, READTYPE, FILTERTYPE], ABC):
    """Abstract base for asyncpg-backed CRUD models.

    Implements shared helpers for parameterized SQL execution and common CRUD
    patterns. Subclasses must provide table/view creation and any
    domain-specific queries.
    """

    # Ensure subclasses have a known logger attribute for type checkers
    logger: logging.Logger

    def __init__(
        self,
        table_name: str,
        db_pool: Pool,
        read_schema: type[READTYPE],
    ) -> None:
        """Initialize model metadata and database connection pool.

        Args:
            table_name: Backing relation name (table or view).
            db_pool: Asyncpg connection pool.
            read_schema: Pydantic model used to validate/read rows.
        """
        self.name = table_name
        self.read_schema = read_schema
        self.db_pool = db_pool
        # Emit debug logs in development mode
        level = LogLevel.DEBUG if settings.arch_stats_dev_mode else LogLevel.INFO
        self.logger = get_logger(__name__, level)

    async def execute(self, sql_statement: str, values: ValuesList | None = None) -> int:
        """Execute a single SQL statement.

        Args:
            sql_statement: SQL command to run (parameterized, no interpolation).
            values: Optional positional parameters to bind.

        Returns:
            Number of rows affected as reported by asyncpg.

        Raises:
            DBException: If execution fails for any reason.
        """
        async with self.db_pool.acquire() as conn:
            try:
                if values is None or not values:
                    result = await conn.execute(sql_statement)
                else:
                    result = await conn.execute(sql_statement, *values)
            except Exception as e:
                raise DBException(e) from e
        # result is a string like 'UPDATE 1' or 'DELETE 0' or 'INSERT 1'
        try:
            affected = int(result.split()[-1])
        except (IndexError, ValueError):
            affected = 0  # default to 0 if parsing fails
        return affected

    async def create(self) -> None:
        """Create the backing relation for this model (table or view).

        Raises:
            NotImplementedError: Must be implemented by subclasses.
        """
        raise NotImplementedError("Error: create method not implemented in child class")

    async def drop(self) -> None:
        """Drop the backing relation for this model (table or view).

        Raises:
            NotImplementedError: Must be implemented by subclasses.
        """
        raise NotImplementedError("Error: drop method not implemented in child class")

    async def check_table_exists(self) -> bool:
        """Check whether the backing table exists in the current search_path.

        Returns:
            True if the relation can be resolved via to_regclass, else False.

        Raises:
            DBException: If the check fails.
        """
        async with self.db_pool.acquire() as conn:
            try:
                regclass = await conn.fetchval("SELECT to_regclass($1);", self.name)
            except Exception as e:
                # On any error, conservatively return False and let caller decide next steps
                raise DBException(e) from e
        return regclass is not None

    async def insert_one(self, data: CREATETYPE) -> UUID:
        """Insert a single record.

        Args:
            data: Pydantic model containing columns and values to insert.

        Returns:
            UUID of the newly inserted record.

        Raises:
            DBException: If insertion fails or no id is returned.
        """

        dump = data.model_dump(by_alias=True, exclude_unset=True)
        keys = list(dump.keys())
        values = list(dump.values())
        columns = ", ".join(keys)
        placeholders = ", ".join(f"${i+1}" for i in range(len(keys)))
        sql_statement = f"INSERT INTO {self.name} ({columns}) VALUES ({placeholders}) RETURNING id;"
        try:
            async with self.db_pool.acquire() as conn:
                row = await conn.fetchrow(sql_statement, *values)
            if not row or "id" not in row or not isinstance(row["id"], UUID):
                raise DBException("Insert failed to return id")
            _id = row["id"]
        except Exception as e:
            raise DBException(e) from e
        return _id

    async def delete_one(self, _id: UUID) -> None:
        """Delete a single record by id.

        Args:
            _id: Record identifier.

        Raises:
            DBException: If the provided id is invalid or execution fails.
            DBNotFound: If no record matches the given id.
        """
        if not _id or not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        sql_statement = f"DELETE FROM {self.name} WHERE id = $1;"
        affected = await self.execute(sql_statement, [_id])
        if affected == 0:
            raise DBNotFound(f"{self.name}: No record found with id={_id}")

    async def update_one(self, _id: UUID, data: UPDATETYPE) -> None:
        """Update a single record by id.

        Args:
            _id: Record identifier.
            data: Pydantic model with fields to update (partial allowed).

        Raises:
            DBNotFound: If no record matches the given id.
        """

        dump = data.model_dump(by_alias=True, exclude_unset=True)
        keys = list(dump.keys())
        values = list(dump.values())
        set_clause = ", ".join(f"{key} = ${i+1}" for i, key in enumerate(keys))
        # The id is the last parameter
        sql_statement = f"UPDATE {self.name} SET {set_clause} WHERE id = ${len(keys)+1};"
        values.append(_id)
        affected = await self.execute(sql_statement, values)
        if affected == 0:
            raise DBNotFound(f"{self.name}: No record found with id={_id}")

    async def get_one(self, where: FILTERTYPE) -> READTYPE:
        """Fetch a single record matching provided filters.

        Args:
            where: Pydantic filter model (only set/non-null fields are applied).

        Returns:
            Instance of the configured read schema.

        Raises:
            DBNotFound: If no record matches the filters.
        """
        select_stm = f"SELECT * FROM {self.name}"
        dump = where.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        values: list[Values] = []
        if dump:
            keys = list(dump.keys())
            values = list(dump.values())
            conditions = " AND ".join(f"{key} = ${i+1}" for i, key in enumerate(keys))
            select_stm += f" WHERE {conditions}"
        select_stm += ";"
        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(select_stm, *values)
            if not row:
                raise DBNotFound(f"{self.name}: No record found with {where=}")
            return self.read_schema(**row)

    async def get_one_by_id(self, _id: UUID) -> READTYPE:
        """Fetch a single record by id.

        Subclasses may override to add constraints or joins.

        Raises:
            NotImplementedError: Must be implemented by subclasses.
        """
        raise NotImplementedError("Error: get_one_by_id method not implemented in child class")

    async def get_all(self, where: FILTERTYPE | None = None) -> list[READTYPE]:
        """Fetch all records matching optional filters.

        Args:
            where: Optional filter model; if omitted, returns all rows.

        Returns:
            List of validated read-schema instances.
        """
        sql_statement = f"SELECT * FROM {self.name}"
        values: list[Values] | None = None
        if where is not None:
            dump = where.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
            if dump:
                keys = list(dump.keys())
                values = list(dump.values())
                conditions = " AND ".join(f"{key} = ${i+1}" for i, key in enumerate(keys))
                sql_statement += f" WHERE {conditions}"
        sql_statement += ";"
        async with self.db_pool.acquire() as conn:
            rows = (
                await conn.fetch(sql_statement, *values)
                if values
                else await conn.fetch(sql_statement)
            )

        return [self.read_schema(**row) for row in rows]

    async def get_by_session_id(self, session_id: UUID) -> list[READTYPE]:
        """Fetch records scoped by a session id.

        Raises:
            NotImplementedError: Must be implemented by subclasses that support
                session scoping.
        """
        raise NotImplementedError("Error: get_by_session_id method not implemented in child class")
