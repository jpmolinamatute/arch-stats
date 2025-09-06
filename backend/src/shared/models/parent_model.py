from abc import ABC
from datetime import datetime
from typing import Generic, Protocol, TypeVar
from uuid import UUID

from asyncpg import Pool
from pydantic import BaseModel


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
    """
    Base class for database operations using asyncpg.
    Provides basic CRUD methods.
    """

    def __init__(
        self,
        table_name: str,
        db_pool: Pool,
        read_schema: type[READTYPE],
    ) -> None:
        """
        Initialize the ParentModel with table info and connection pool.

        Args:
            table_name: The name of the database table.
            db_pool: The asyncpg connection pool.
            read_schema: The Pydantic model used for reads.
            jsonb_fields: Optional list of column names that are JSONB and need decoding.
        """
        self.name = table_name
        self.read_schema = read_schema
        self.db_pool = db_pool

    async def execute(self, sql_statement: str, values: ValuesList | None = None) -> int:
        """
        Execute an SQL statement with optional parameter binding.

        Args:
            sql_statement: The SQL command to run.
            values: Optional list of values to bind to the SQL placeholders.

        Returns:
            Number of rows affected.
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
        """Create the backing relation for this model (table or view)."""
        raise NotImplementedError("Error: create method not implemented in child class")

    async def drop(self) -> None:
        """Drop the backing relation for this model (table or view)."""
        raise NotImplementedError("Error: drop method not implemented in child class")

    async def check_table_exists(self) -> bool:
        """Return True if the table exists in the current search_path.

        Uses SELECT to_regclass($1) which returns NULL when the relation doesn't
        exist. This must use a fetch operation (not execute) to read the value.
        """
        async with self.db_pool.acquire() as conn:
            try:
                regclass = await conn.fetchval("SELECT to_regclass($1);", self.name)
            except Exception as e:
                # On any error, conservatively return False and let caller decide next steps
                raise DBException(e) from e
        return regclass is not None

    async def insert_one(self, data: CREATETYPE) -> UUID:
        """
        Insert a new record into the database.

        Args:
            data: A CREATETYPE Pydantic model containing the data to insert.

        Returns:
            The UUID of the newly inserted record.
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
        """
        Delete a record by its UUID.

        Args:
            _id: The UUID of the record to delete.

        Raises:
            DBNotFound if no record is found.
        """
        if not _id or not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        sql_statement = f"DELETE FROM {self.name} WHERE id = $1;"
        affected = await self.execute(sql_statement, [_id])
        if affected == 0:
            raise DBNotFound(f"{self.name}: No record found with id={_id}")

    async def update_one(self, _id: UUID, data: UPDATETYPE) -> None:
        """
        Update a record by its UUID.

        Args:
            _id: The UUID of the record.
            data: A UPDATETYPE Pydantic model with the update data.

        Raises:
            DBNotFound if no record is found.
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
        """
        Retrieve a single record that matches the provided filters.

        Args:
            where: Dictionary of column filters.

        Returns:
            A READTYPE instance.

        Raises:
            DBNotFound if no matching record is found.
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
        raise NotImplementedError("Error: get_one_by_id method not implemented in child class")

    async def get_all(self, where: FILTERTYPE | None = None) -> list[READTYPE]:
        """
        Retrieve all records matching the optional filters.

        Args:
            where: Optional dictionary of filters.

        Returns:
            List of READTYPE instances.
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
        raise NotImplementedError("Error: get_by_session_id method not implemented in child class")
