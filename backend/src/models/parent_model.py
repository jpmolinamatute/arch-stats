import logging
from abc import ABC
from collections.abc import Sequence
from datetime import datetime
from typing import Protocol
from uuid import UUID

from asyncpg import Pool, Record
from pydantic import BaseModel

# NOTE: Import get_logger directly from its module to avoid importing the
# core package __init__ (which re-exports SessionManager and other modules),
# preventing cyclic imports with models -> parent_model -> core -> session_manager -> models
from core.logger import get_logger
from models.sql_statement_builder import SQLStatementBuilder

type SimpleValues = str | float | bool | int
type Values = SimpleValues | UUID | datetime | bytes | None | Sequence[int] | Sequence[float]
type ValuesTuple = Sequence[Values]


class HasId(Protocol):
    def get_id(self) -> UUID: ...


class DBException(Exception):
    pass


class DBNotFound(Exception):
    pass


class ParentModel[
    CREATETYPE: BaseModel,
    SETTYPE: BaseModel,
    READTYPE: HasId,
    FILTERTYPE: BaseModel,
](ABC):
    """Abstract base for asyncpg-backed CRUD models.

    Implements shared helpers for parameterized SQL execution and common CRUD
    patterns. Subclasses must provide table/view creation and any
    domain-specific queries.
    """

    # Ensure subclasses have a known logger attribute for type checkers
    logger: logging.Logger

    def __init__(self, table_name: str, db_pool: Pool, read_schema: type[READTYPE]) -> None:
        """Initialize model metadata and database connection pool.

        Args:
            table_name: Backing relation name (table or view).
            db_pool: Asyncpg connection pool.
            read_schema: Pydantic model used to validate/read rows.
        """
        self.name = table_name
        self.read_schema = read_schema
        self.pk = f"{self.name}_id"
        self.db_pool = db_pool
        self.logger = get_logger()
        # SQL builder scoped to this model's primary table. Use for safe SQL assembly.
        self.sql_builder = SQLStatementBuilder(self.name)

    def build_select_sql_stm(
        self,
        where: FILTERTYPE,
        columns: list[str],
        limit: int = 1,
        is_desc: bool = False,
    ) -> tuple[str, ValuesTuple]:
        """Assemble a parameterized SELECT using SQLStatementBuilder.

        Builds conditions from the provided filter model and returns the
        SQL string along with a tuple of values in placeholder order.
        """
        dump = where.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        keys = list(dump.keys())
        values: ValuesTuple = tuple(dump.values()) if dump else ()

        conditions = [f"{key} = ${i}" for i, key in enumerate(keys, start=1)]
        sql_stm = self.sql_builder.build_select_with_conditions(columns, conditions, limit, is_desc)
        return (sql_stm, values)

    def build_update_sql_stm(self, data: SETTYPE, where: FILTERTYPE) -> tuple[str, ValuesTuple]:
        """Assemble a parameterized UPDATE using SQLStatementBuilder.

        Placeholder numbering for WHERE conditions continues after the SET
        clause to match the bound values order.
        """
        set_dump = data.model_dump(by_alias=False, exclude_unset=True, exclude_none=True)
        set_keys = list(set_dump.keys())
        set_values = tuple(set_dump.values())

        set_data = [(key, f"${i}") for i, key in enumerate(set_keys, start=1)]

        where_dump = where.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        where_keys = list(where_dump.keys())
        where_values = tuple(where_dump.values())

        conditions = [f"{key} = ${i}" for i, key in enumerate(where_keys, start=len(set_keys) + 1)]

        sql_statement = self.sql_builder.build_update(set_data, conditions)
        values: ValuesTuple = (*set_values, *where_values)
        return (sql_statement, values)

    def build_insert_sql_stm(self, data: CREATETYPE | list[CREATETYPE]) -> tuple[str, ValuesTuple]:
        """Assemble a parameterized INSERT using SQLStatementBuilder.

        Returns the SQL string along with a tuple of values in placeholder order.
        """
        if isinstance(data, list):
            data_list = data
        else:
            data_list = [data]

        # if data_list is empty, raise ValueError
        if not data_list:
            raise ValueError("No data provided for insert")

        # Assume all items have the same keys (Pydantic ensures this for same model type)
        first_dump = data_list[0].model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        keys = list(first_dump.keys())

        all_values: list[Values] = []
        for item in data_list:
            dump = item.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
            # Ensure order matches keys
            for key in keys:
                all_values.append(dump[key])

        sql_stm = self.sql_builder.build_insert(keys, num_rows=len(data_list))
        return (sql_stm, all_values)

    def build_delete_sql_stm(self, where: FILTERTYPE) -> tuple[str, ValuesTuple]:
        """Assemble a parameterized DELETE using SQLStatementBuilder."""
        dump = where.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        keys = list(dump.keys())
        values = tuple(dump.values()) if dump else ()

        conditions = [f"{key} = ${i}" for i, key in enumerate(keys, start=1)]
        sql_stm = self.sql_builder.build_delete(conditions)
        return (sql_stm, values)

    def build_select_view_sql_stm(
        self,
        view_name: str,
        where: BaseModel,
        columns: list[str],
        limit: int = 1,
        is_desc: bool = False,
    ) -> tuple[str, ValuesTuple]:
        """Assemble a parameterized SELECT from a VIEW using SQLStatementBuilder."""
        dump = where.model_dump(by_alias=True, exclude_unset=True, exclude_none=True)
        keys = list(dump.keys())
        values = tuple(dump.values()) if dump else ()

        conditions = [f"{key} = ${i}" for i, key in enumerate(keys, start=1)]
        order_by = "ORDER BY created_at DESC" if is_desc else ""

        sql_stm = self.sql_builder.build_select_view(
            view_name=view_name,
            columns=columns,
            conditions=conditions,
            order_by=order_by,
            limit=limit,
        )
        return (sql_stm, values)

    def build_select_function_sql_stm(
        self, function_name: str, args: list[Values]
    ) -> tuple[str, ValuesTuple]:
        """Assemble a parameterized SELECT from a FUNCTION using SQLStatementBuilder."""
        sql_stm = self.sql_builder.build_select_function(function_name, len(args))
        return (sql_stm, tuple(args))

    async def fetch(self, query_data: tuple[str, ValuesTuple]) -> list[Record]:
        """Execute a SELECT query and return a list of records.

        Args:
            query_data: Tuple containing the SQL statement and values.

        Returns:
            List of asyncpg.Record objects. Returns empty list if no results.
        """
        sql_statement, values = query_data
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", sql_statement)
            rows = await conn.fetch(sql_statement, *values)
        return rows

    async def fetchrow(self, query_data: tuple[str, ValuesTuple]) -> Record:
        """Execute a SELECT query and return a single record.

        Args:
            query_data: Tuple containing the SQL statement and values.

        Returns:
            asyncpg.Record object.

        Raises:
            DBNotFound: If no record is found.
        """
        sql_statement, values = query_data
        async with self.db_pool.acquire() as conn:
            self.logger.debug("Fetching: %s", sql_statement)
            row = await conn.fetchrow(sql_statement, *values)
        if not row:
            raise DBNotFound(f"{self.name}: No record found")
        return row

    async def execute(self, sql_statement: str, values: ValuesTuple | None = None) -> int:
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
                self.logger.debug("Executing SQL: %s", sql_statement)
                if values is None or not values:
                    result = await conn.execute(sql_statement)
                else:
                    result = await conn.execute(sql_statement, *values)
            except Exception as e:
                raise DBException(e) from e
        # result is a string like 'UPDATE 1' or 'DELETE 0' or 'INSERT 1'
        try:
            affected = int(result.split()[-1])
            affected
        except (IndexError, ValueError):
            affected = 0  # default to 0 if parsing fails
        return affected

    async def insert_one(self, data: CREATETYPE) -> UUID:
        """Insert a single record.

        Args:
            data: Pydantic model containing columns and values to insert.

        Returns:
            UUID of the newly inserted record.

        Raises:
            DBException: If insertion fails or no id is returned.
        """

        sql_statement, values = self.build_insert_sql_stm(data)
        try:
            row = await self.fetchrow((sql_statement, values))
            if self.pk not in row or not isinstance(row[self.pk], UUID):
                raise DBException(f"Insert failed to return {self.pk}")
            _id: UUID = row[self.pk]
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
        sql_statement = f"DELETE FROM {self.name} WHERE {self.pk} = $1;"
        affected = await self.execute(sql_statement, (_id,))
        if affected == 0:
            raise DBNotFound(f"{self.name}: No record found with {self.pk}={_id}")

    async def get_one(self, where: FILTERTYPE) -> READTYPE:
        """Fetch a single record matching provided filters.

        Args:
            where: Pydantic filter model (only set/non-null fields are applied).

        Returns:
            Instance of the configured read schema.

        Raises:
            DBNotFound: If no record matches the filters.
        """
        query_data = self.build_select_sql_stm(where, [], 1, False)
        row = await self.fetchrow(query_data)
        return self.read_schema(**row)

    async def update(self, data: SETTYPE, where: FILTERTYPE) -> None:
        """Update a single record by id.

        Args:
            _id: Record identifier.
            data: Pydantic model with fields to update (partial allowed).

        Raises:
            DBNotFound: If no record matches the given id.
        """
        sql_statement, values = self.build_update_sql_stm(data, where)
        affected = await self.execute(sql_statement, values)
        if affected == 0:
            raise DBNotFound(f"{self.name}: No record(s) found")

    async def insert_many(self, data: list[CREATETYPE]) -> list[UUID]:
        """Insert multiple records at once."""
        sql_statement, values = self.build_insert_sql_stm(data)

        rows = await self.fetch((sql_statement, values))

        if not rows:
            raise DBNotFound(f"No {self.name} created")

        return [r[self.pk] for r in rows]

    async def get_all(self, where: FILTERTYPE, columns: list[str]) -> list[READTYPE]:
        """Fetch all records matching optional filters.

        Args:
            where: Optional filter model; if omitted, returns all rows.

        Returns:
            List of validated read-schema instances.
        """
        query_data = self.build_select_sql_stm(where, columns, 0, False)
        rows = await self.fetch(query_data)

        return [self.read_schema(**row) for row in rows]

    async def get_by_session_id(self, session_id: UUID) -> list[READTYPE]:
        """Fetch records scoped by a session id.

        Raises:
            NotImplementedError: Must be implemented by subclasses that support
                session scoping.
        """
        raise NotImplementedError("Error: get_by_session_id method not implemented in child class")
