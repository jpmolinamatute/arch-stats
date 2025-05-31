import logging
import sys
from enum import IntEnum
from logging.handlers import TimedRotatingFileHandler
from logging import StreamHandler, Formatter, Handler
from pathlib import Path


class LogLevel(IntEnum):
    CRITICAL = 50
    ERROR = 40
    WARNING = 30
    INFO = 20
    DEBUG = 10


def create_file_handler(formatter: Formatter, log_lever: LogLevel) -> Handler:
    log_dir = Path(__file__).parent.joinpath("logs/")
    log_dir.mkdir(exist_ok=True)
    log_file = log_dir.joinpath("app.log")

    file_handler = TimedRotatingFileHandler(
        filename=log_file,
        when="midnight",  # Rotate logs every midnight
        interval=1,  # Every 1 day
        backupCount=7,  # Keep 7 days of logs
        encoding="utf-8",  # Optional: ensures compatibility
        utc=True,  # Use UTC for rotation time (set False for local time)
    )
    file_handler.setLevel(log_lever)
    file_handler.setFormatter(formatter)

    return file_handler


def create_stream_handler(formatter: Formatter, log_lever: LogLevel) -> Handler:
    console_handler = StreamHandler(sys.stdout)
    console_handler.setLevel(log_lever)
    console_handler.setFormatter(formatter)
    return console_handler


def get_logger(file_name: str, log_lever: LogLevel) -> logging.Logger:
    logger = logging.getLogger(file_name)
    logger.setLevel(log_lever)
    if not logger.hasHandlers():
        formatter = Formatter(
            "%(asctime)s|%(pathname)s:%(lineno)d|%(message)s", datefmt="%Y-%m-%d %H:%M:%S"
        )

        file_handler = create_file_handler(formatter, log_lever)
        console_handler = create_stream_handler(formatter, log_lever)

        logger.addHandler(file_handler)
        logger.addHandler(console_handler)

    return logger
