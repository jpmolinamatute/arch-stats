#!/usr/bin/env python

import asyncio
import random
from datetime import datetime
from logging import Logger
from os import getenv
from pathlib import Path
from uuid import UUID

from asyncpg import Connection, connect

from shared import SensorDataTuple


class SensorDataError(Exception):
    """Custom exception for sensor data errors."""


def fake_sensor_data(hardcoded: bool | None = True) -> bool:
    """Simulate sensor data availability."""
    if hardcoded is not None:
        choice = hardcoded
    else:
        choice = random.choice([True, False])

    return choice


def get_arrow_engage_time() -> datetime:
    """Get the arrow engage time from the sensor."""
    if fake_sensor_data():
        return datetime.now()
    raise SensorDataError("Arrow engage time not found")


def get_draw_length() -> float:
    if fake_sensor_data():
        return 0.0
    raise SensorDataError("Draw length not found")


def get_arrow_disengage_time() -> datetime:
    if fake_sensor_data():
        return datetime.now()
    raise SensorDataError("Arrow disengage time not found")


def get_arrow_landing_time() -> datetime | None:
    if fake_sensor_data():
        return datetime.now()
    raise SensorDataError("Arrow landing time not found")


def get_x_coordinate() -> float | None:
    if fake_sensor_data():
        return 0.0
    raise SensorDataError("x coordinate not found")


def get_y_coordinate() -> float | None:
    if fake_sensor_data():
        return 0.0
    raise SensorDataError("y coordinate not found")


def get_distance() -> float:
    if fake_sensor_data():
        return 0.0
    raise SensorDataError("distance not found")


def get_arrow_id() -> UUID:
    list_of_ids = [
        "0dc6c7b5-f584-4538-82f3-c0081a6596a0",
        "96458074-a2a9-44e0-87f3-4e3188da5a36",
        "b0deea9c-ca98-4ee4-9a64-cd4631ce9f5a",
        "93ea2707-eff7-4167-b2b3-ecb260919c5b",
        "982757d6-d704-408c-9faf-67e600441d6c",
        "17e212aa-321b-4dd8-b3dc-55c513f07f9e",
        "527b6cf8-de95-48d1-9137-656c173a5373",
        "49c16351-4f01-4589-aa48-f9eecc1fb0fc",
        "8ce33679-e0b0-4a96-ab5a-e4fbb9319b2d",
        "1d4c8108-cc40-4731-b85c-503d56ade888",
    ]
    return UUID(random.choice(list_of_ids))


def get_target_track_id() -> UUID:
    arch_dir = getenv("ARCH_STATS_DIR")
    arch_id_file = getenv("ARCH_STATS_ID_FILE")
    if arch_dir is None or arch_id_file is None:
        raise ValueError("ARCH_STATS_DIR and ARCH_STATS_ID_FILE must be set")
    id_file = Path(f"{arch_dir}/backend/{arch_id_file}")
    target_track_id: UUID
    if id_file.exists():
        with open(id_file, "r", encoding="utf-8") as file:
            target_track_id = UUID(file.read().strip())
    else:
        raise FileNotFoundError(f"File {id_file} not found")
    return target_track_id


def read_all_sensor_data(target_track_id: UUID) -> SensorDataTuple:
    return (
        target_track_id,
        get_arrow_id(),
        get_arrow_engage_time(),
        get_draw_length(),
        get_arrow_disengage_time(),
        get_arrow_landing_time(),
        get_x_coordinate(),
        get_y_coordinate(),
        get_distance(),
    )


async def save_sensor_data(conn: Connection, target_track_id: UUID, logger: Logger) -> None:
    """Save sensor data to the database."""
    try:
        all_data = read_all_sensor_data(target_track_id)
    except SensorDataError:
        logger.error("Failed to read sensor data")
    else:
        logger.info("Inserting sensor data into the database")
        logger.info(all_data)
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
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
        """
        await conn.execute(insert_stm, *all_data)


async def setup(logger: Logger) -> None:
    logger.info("Starting the shoot_recorder...")
    target_track_id = get_target_track_id()
    user = getenv("ARCH_STATS_USER")
    params = {}
    if user:
        params["user"] = user

    conn = await connect(**params)
    try:
        while True:
            await save_sensor_data(conn, target_track_id, logger)
            await asyncio.sleep(5)
    finally:
        await conn.close()
