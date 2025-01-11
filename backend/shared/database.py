from __future__ import annotations
from os import getenv
from typing import Any, Type
from types import TracebackType

import psycopg2
import psycopg2.extras
from psycopg2.extensions import connection
from shared.typed import SensorDataTuple


class DataBaseError(Exception):
    pass


class DataBase:
    def __init__(self) -> None:
        self.conn = self.connect_to_db()
        psycopg2.extras.register_uuid(conn_or_curs=self.conn)

    def connect_to_db(self) -> connection:
        user = getenv("ARCH_STATS_USER")
        if user:
            conn = psycopg2.connect(user=user)
        else:
            conn = psycopg2.connect()
        with conn.cursor() as cursor:
            cursor.execute("SELECT 1")
        return conn

    def close(self) -> None:
        self.conn.close()

    def __enter__(self) -> DataBase:
        return self

    def __exit__(
        self,
        exc_type: Type[BaseException] | None,
        exc_value: BaseException | None,
        _: TracebackType | None,
    ) -> None:
        self.close()
        if exc_type is KeyboardInterrupt:
            raise KeyboardInterrupt
        elif exc_type is not None:
            raise DataBaseError("Database error occurred") from exc_value

    def insert(self, query: str, data: tuple[Any, ...]) -> None:
        with self.conn.cursor() as cursor:
            cursor.execute(query, data)
            self.conn.commit()

    def query(self, query: str) -> list[tuple[Any, ...]]:
        with self.conn.cursor() as cursor:
            cursor.execute(query)
            result: list[tuple[Any, ...]] = cursor.fetchall()
            return result

    def insert_shooting(self, data: SensorDataTuple) -> None:
        insert_stm = """
            INSERT INTO shooting (
                target_track_id,
                arrow_id,
                arrow_engage_time,
                draw_length,
                arrow_disengage_time,
                arrow_landing_time,
                x_coordinate,
                y_coordinate,
                distance
            )
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s);
        """
        self.insert(insert_stm, data)
