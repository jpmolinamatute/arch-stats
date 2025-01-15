from __future__ import annotations
import json
from os import getenv
from typing import Any, Callable

import select
import psycopg

from shared.typed import SensorData
from shared.logger import get_logger


class DataBaseError(Exception):
    pass


class DataBase:
    def __init__(self) -> None:
        self.logger = get_logger()
        self.conn = self.connect_to_db()

    def connect_to_db(self) -> psycopg.Connection:
        self.logger.info("Creating a new connection to the database")
        user = getenv("ARCH_STATS_USER")
        if user:
            conn = psycopg.connect(user=user)
        else:
            conn = psycopg.connect()
        with conn.cursor() as cursor:
            cursor.execute("SELECT 1")
        return conn

    def close(self) -> None:
        self.logger.info("Closing a connection to the database")
        self.conn.close()

    def query(self, query: str) -> list[tuple[Any, ...]]:
        with self.conn.cursor() as cursor:
            cursor.execute(query)
            result: list[tuple[Any, ...]] = cursor.fetchall()
            return result

    # def watch_shooting_change(self, callback: Callable[[SensorData], None]) -> None:
    # with self.conn.cursor() as cursor:
    #     cursor.execute("LISTEN shooting_change;")
    #     self.conn.commit()
    #         while True:
    #             if select.select([self.conn], [], [], 5) == ([], [], []):
    #                 continue
    #             self.conn.poll()
    #             while self.conn.notifies:
    #                 notify = self.conn.notifies.pop(0)
    #                 payload = json.loads(notify.payload)
    #                 self.logger.info("Received notification: %s", payload)
    #                 callback(payload)
