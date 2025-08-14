import json
from abc import ABC
from datetime import datetime
from typing import Generic, Protocol, TypeVar
from uuid import UUID

from asyncpg import Pool, Record
from pydantic import BaseModel


Values = str | float | bool | int | UUID | datetime | None | list[int] | list[float]
DictValues = dict[str, Values]
ValuesList = list[Values]


class HasId(Protocol):
    def get_id(self) -> UUID: ...


CREATETYPE = TypeVar("CREATETYPE", bound=BaseModel)
UPDATETYPE = TypeVar("UPDATETYPE", bound=BaseModel)
READTYPE = TypeVar("READTYPE", bound=HasId)


class DBException(Exception):
    pass


class DBNotFound(Exception):
    pass


class DBBase(Generic[CREATETYPE, UPDATETYPE, READTYPE], ABC):
    """
    Base class for database operations using asyncpg.
    Provides basic CRUD methods.
    """

    def __init__(
        self,
        table_name: str,
        table_schema: str,
        db_pool: Pool,
        read_schema: type[READTYPE],
        jsonb_fields: list[str] | None = None,
    ) -> None:
        """
        Initialize the DBBase with table info and connection pool.

        Args:
            table_name: The name of the database table.
            table_schema: The SQL schema definition for the table.
            db_pool: The asyncpg connection pool.
            read_schema: The Pydantic model used for reads.
            jsonb_fields: Optional list of column names that are JSONB and need decoding.
        """
        self.table_name = table_name
        self.table_schema = table_schema
        self.read_schema = read_schema
        self.db_pool = db_pool
        self.jsonb_fields = jsonb_fields

    def _convert_row(self, row: Record) -> READTYPE:
        """
        Convert a database row to a READTYPE instance, decoding any JSONB fields.

        Args:
            row: The asyncpg Record from a SELECT query.

        Returns:
            An instance of READTYPE with JSONB fields deserialized.
        """
        row_dict = dict(row)
        if self.jsonb_fields:
            for column in self.jsonb_fields:
                if column in row_dict and isinstance(row_dict[column], str):
                    try:
                        row_dict[column] = json.loads(row_dict[column])
                    except json.JSONDecodeError as e:
                        raise DBException(f"Failed to decode '{column}' JSON string: {e}") from e
        return self.read_schema(**row_dict)

    def _serialize_db_value(self, values: Values) -> Values:
        """
        Convert complex types like dicts/lists to JSON strings for database binding.

        Args:
            values: The value to convert.

        Returns:
            The value ready for DB binding.
        """
        known_type_value: Values
        if isinstance(values, (dict, list)):
            known_type_value = json.dumps(values)
        else:
            known_type_value = values
        return known_type_value

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

    async def create_table(self) -> None:
        """
        Create the table using the provided schema if it does not exist.
        """
        sql_statement = f"CREATE TABLE IF NOT EXISTS {self.table_name} ({self.table_schema});"
        await self.execute(sql_statement)

    async def drop_table(self) -> None:
        """
        Drop the table from the database if it exists.
        """
        sql_statement = f"DROP TABLE IF EXISTS {self.table_name};"
        await self.execute(sql_statement)

    async def insert_one(self, data: CREATETYPE) -> UUID:
        """
        Insert a new record into the database.

        Args:
            data: A CREATETYPE Pydantic model containing the data to insert.

        Returns:
            The UUID of the newly inserted record.
        """

        data_dict = data.model_dump(by_alias=True, exclude_unset=True)
        if not data_dict:
            raise DBException("Error: invalid 'data' provided to insert_one method.")
        keys = list(data_dict.keys())
        values = [self._serialize_db_value(v) for v in data_dict.values()]
        columns = ", ".join(keys)
        placeholders = ", ".join(f"${i+1}" for i in range(len(keys)))
        sql_statement = f"INSERT INTO {self.table_name} ({columns}) VALUES ({placeholders});"
        try:
            await self.execute(sql_statement, values)
            row = await self.get_one(data_dict)
            _id = row.get_id()
            if not isinstance(_id, UUID):
                raise DBException("Insert failed to return id")
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
        sql_statement = f"DELETE FROM {self.table_name} WHERE id = $1;"
        affected = await self.execute(sql_statement, [_id])
        if affected == 0:
            raise DBNotFound(f"{self.table_name}: No record found with id={_id}")

    async def update_one(self, _id: UUID, data: UPDATETYPE) -> None:
        """
        Update a record by its UUID.

        Args:
            _id: The UUID of the record.
            data: A UPDATETYPE Pydantic model with the update data.

        Raises:
            DBNotFound if no record is found.
        """

        data_dict = data.model_dump(exclude_unset=True)
        if not data_dict:
            raise DBException("Error: invalid 'data' provided to update_one method.")
        keys = list(data_dict.keys())
        values = list(data_dict.values())
        set_clause = ", ".join(f"{key} = ${i+1}" for i, key in enumerate(keys))
        # The id is the last parameter
        sql_statement = f"UPDATE {self.table_name} SET {set_clause} WHERE id = ${len(keys)+1};"
        values.append(_id)
        affected = await self.execute(sql_statement, values)
        if affected == 0:
            raise DBNotFound(f"{self.table_name}: No record found with id={_id}")

    async def get_one(self, where: DictValues) -> READTYPE:
        """
        Retrieve a single record that matches the provided filters.

        Args:
            where: Dictionary of column filters.

        Returns:
            A READTYPE instance.

        Raises:
            DBNotFound if no matching record is found.
        """
        select_stm = f"""
            SELECT *
            FROM {self.table_name}
            WHERE
        """
        conditions = " AND ".join(f"{key} = ${i+1}" for i, key in enumerate(where.keys()))
        select_stm += f"{conditions};"
        values = [self._serialize_db_value(v) for v in where.values()]
        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(select_stm, *values)
            if not row:
                raise DBNotFound(f"{self.table_name}: No record found with {where=}")
            return self._convert_row(row)

    async def get_one_by_id(self, _id: UUID) -> READTYPE:
        """
        Retrieve a single record by its UUID.

        Args:
            _id: The UUID of the record.

        Returns:
            A READTYPE instance.
        """
        if not _id or not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        return await self.get_one({"id": _id})

    async def get_all(self, where: DictValues | None = None) -> list[READTYPE]:
        """
        Retrieve all records matching the optional filters.

        Args:
            where: Optional dictionary of filters.

        Returns:
            List of READTYPE instances.
        """
        sql_statement = f"SELECT * FROM {self.table_name}"
        if where:
            conditions = " AND ".join(f"{key} = ${i+1}" for i, key in enumerate(where.keys()))
            sql_statement += f" WHERE {conditions};"
            values = [self._serialize_db_value(v) for v in where.values()]
        else:
            values = None
            sql_statement += ";"
        async with self.db_pool.acquire() as conn:
            rows = (
                await conn.fetch(sql_statement, *values)
                if values
                else await conn.fetch(sql_statement)
            )

        return [self._convert_row(row) for row in rows]

    async def get_by_session_id(self, session_id: UUID) -> list[READTYPE]:
        raise NotImplementedError("Error: get_by_session_id method not implemented in child class")
