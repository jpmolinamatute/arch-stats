from __future__ import annotations

import logging
import sys
from logging import Formatter, Handler, StreamHandler
from logging.handlers import TimedRotatingFileHandler
from pathlib import Path
from threading import Lock
from typing import ClassVar, Literal

from core.settings import settings


# Narrow integer type for log levels
type LogLevel = Literal[0, 10, 20, 30, 40, 50]


class LoggerFactory:
    """Singleton that configures and returns loggers and handlers.

    Uses TimedRotatingFileHandler for file logs and a stream handler for stdout.
    """

    _instance: ClassVar[LoggerFactory | None] = None
    _lock: ClassVar[Lock] = Lock()

    def __new__(cls) -> LoggerFactory:
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
        assert cls._instance is not None
        return cls._instance

    def create_file_handler(self, formatter: Formatter, log_level: LogLevel) -> Handler:
        """Create a timed rotating file handler.

        Args:
            formatter: Formatter to apply to the handler.
            log_level: Minimum log level for this handler.

        Returns:
            Configured file handler.
        """
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
        file_handler.setLevel(int(log_level))
        file_handler.setFormatter(formatter)
        return file_handler

    def create_stream_handler(self, formatter: Formatter, log_level: LogLevel) -> Handler:
        """Create a stdout stream handler.

        Args:
            formatter: Formatter to apply to the handler.
            log_level: Minimum log level for this handler.

        Returns:
            Configured stream handler.
        """
        console_handler = StreamHandler(sys.stdout)
        console_handler.setLevel(int(log_level))
        console_handler.setFormatter(formatter)
        return console_handler

    def get_logger(self, file_name: str) -> logging.Logger:
        """Get or create a configured logger.

        The logger is configured once (handlers added once). Subsequent calls
        will adjust the logger level but avoid duplicating handlers.

        Args:
            file_name: Logger name, typically __name__ or module path.
            log_level: Log level for the logger.

        Returns:
            A configured logging.Logger instance.
        """
        logger = logging.getLogger(file_name)

        log_level: LogLevel = 10 if settings.arch_stats_dev_mode else 20
        logger.setLevel(int(log_level))
        if not logger.hasHandlers():
            formatter = Formatter(
                "%(asctime)s|%(pathname)s:%(lineno)d|%(message)s",
                datefmt="%Y-%m-%d %H:%M:%S",
            )
            file_handler = self.create_file_handler(formatter, log_level)
            console_handler = self.create_stream_handler(formatter, log_level)
            logger.addHandler(file_handler)
            logger.addHandler(console_handler)
        return logger
