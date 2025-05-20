from __future__ import annotations

from abc import ABC
from uuid import UUID

from asyncpg import Pool

Values = str | float | bool | int | UUID
BaseData = dict[str, Values]
ValuesList = list[Values]


class DBBase(ABC):
    """
    Base class for database operations using asyncpg.

    Provides basic CRUD methods.
    """

    def __init__(self, table_name: str, table_schema: str, db_pool: Pool) -> None:
        self.table_name = table_name
        self.table_schema = table_schema
        self.db_pool = db_pool

    async def execute(self, sql_statement: str, values: ValuesList | None = None) -> None:
        # assert DBBase.db_pool is not None, "Database pool is not initialized."
        async with self.db_pool.acquire() as conn:
            if values is None:
                await conn.execute(sql_statement)
            else:
                await conn.execute(sql_statement, *values)

    async def create_table(self) -> None:
        """
        Create the table in the database.
        """
        sql_statement = f"CREATE TABLE IF NOT EXISTS {self.table_name} ({self.table_schema});"
        await self.execute(sql_statement)

    async def insert_one(self, data: BaseData) -> None:
        """
        Insert a new record into the table.
        """
        keys = list(data.keys())
        values = list(data.values())
        columns = ", ".join(keys)
        placeholders = ", ".join(f"${i+1}" for i in range(len(keys)))
        sql_statement = f"INSERT INTO {self.table_name} ({columns}) VALUES ({placeholders});"
        await self.execute(sql_statement, values)

    async def delete_one(self, _id: UUID) -> None:
        """
        Delete a record from the table.
        """
        sql_statement = f"DELETE FROM {self.table_name} WHERE id = $1;"
        await self.execute(sql_statement, [_id])

    async def update_one(self, _id: UUID, data: BaseData) -> None:
        """
        Update a record in the table.
        """
        keys = list(data.keys())
        values = list(data.values())
        set_clause = ", ".join(f"{key} = ${i+1}" for i, key in enumerate(keys))
        # The id is the last parameter
        sql_statement = f"UPDATE {self.table_name} SET {set_clause} WHERE id = ${len(keys)+1};"
        values.append(_id)
        await self.execute(sql_statement, values)

    async def get_one(self, _id: UUID) -> BaseData | None:
        """
        Retrieve a record from the table.
        """
        sql_statement = f"SELECT * FROM {self.table_name} WHERE id = $1;"
        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(sql_statement, _id)
            return dict(row) if row else None

    async def get_all(self) -> list[BaseData]:
        """
        Retrieve all records from the table.
        """
        sql_statement = f"SELECT * FROM {self.table_name};"
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(sql_statement)
            return [dict(row) for row in rows]
