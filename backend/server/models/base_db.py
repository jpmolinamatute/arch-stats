from abc import ABC
from datetime import datetime
from typing import Generic, Protocol, TypeVar
from uuid import UUID

from asyncpg import Pool
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
        self, table_name: str, table_schema: str, db_pool: Pool, read_schema: type[READTYPE]
    ) -> None:
        self.table_name = table_name
        self.table_schema = table_schema
        self.read_schema = read_schema
        self.db_pool = db_pool

    async def execute(self, sql_statement: str, values: ValuesList | None = None) -> int:

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
        Create the table in the database.
        """
        sql_statement = f"CREATE TABLE IF NOT EXISTS {self.table_name} ({self.table_schema});"
        await self.execute(sql_statement)

    async def drop_table(self) -> None:
        """
        Drop the table from the database.
        """
        sql_statement = f"DROP TABLE IF EXISTS {self.table_name};"
        await self.execute(sql_statement)

    async def insert_one(self, data: CREATETYPE) -> UUID:
        """
        Insert a new record into the table.
        """

        data_dict = data.model_dump(by_alias=True, exclude_unset=True)
        if not data_dict:
            raise DBException("Error: invalid 'data' provided to insert_one method.")
        keys = list(data_dict.keys())
        values = list(data_dict.values())
        columns = ", ".join(keys)
        placeholders = ", ".join(f"${i+1}" for i in range(len(keys)))
        sql_statement = f"INSERT INTO {self.table_name} ({columns}) VALUES ({placeholders});"
        await self.execute(sql_statement, values)
        try:
            row = await self.get_by_filters(data_dict)
            _id = row.get_id()
            if not isinstance(_id, UUID):
                raise DBException("Insert failed to return id")
        except Exception as e:
            raise DBException(e) from e
        return _id

    async def delete_one(self, _id: UUID) -> None:
        """
        Delete a record from the table.
        """
        if not _id or not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        sql_statement = f"DELETE FROM {self.table_name} WHERE id = $1;"
        affected = await self.execute(sql_statement, [_id])
        if affected == 0:
            raise DBNotFound(f"{self.table_name}: No record found with id={_id}")

    async def update_one(self, _id: UUID, data: UPDATETYPE) -> None:
        """
        Update a record in the table.
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

    async def get_by_filters(self, where: DictValues) -> READTYPE:
        """
        Retrieve a record from the table.
        """
        select_stm = f"""
            SELECT *
            FROM {self.table_name}
            WHERE
        """
        conditions = " AND ".join(f"{key} = ${i+1}" for i, key in enumerate(where.keys()))
        select_stm += f"{conditions};"
        values = list(where.values())
        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(select_stm, *values)
            if not row:
                raise DBNotFound(f"{self.table_name}: No record found with {where=}")
            return self.read_schema(**dict(row))

    async def get_one_by_id(self, _id: UUID) -> READTYPE:
        """
        Retrieve a record from the table by its ID.
        This is an alias for get_one.
        """
        if not _id or not isinstance(_id, UUID):
            raise DBException("Error: invalid '_id' provided to delete_one method.")
        return await self.get_by_filters({"id": _id})

    async def get_all(self, where_stm: DictValues | None = None) -> list[READTYPE]:
        """
        Retrieve all records from the table.
        """
        sql_statement = f"SELECT * FROM {self.table_name}"
        if where_stm:
            conditions = " AND ".join(f"{key} = ${i+1}" for i, key in enumerate(where_stm.keys()))
            sql_statement += f" WHERE {conditions};"
            values = list(where_stm.values())
        else:
            values = None
            sql_statement += ";"
        async with self.db_pool.acquire() as conn:
            rows = (
                await conn.fetch(sql_statement, *values)
                if values
                else await conn.fetch(sql_statement)
            )
            return [self.read_schema(**dict(row)) for row in rows]
