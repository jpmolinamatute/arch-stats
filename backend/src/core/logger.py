import logging
import sys
from logging import Formatter, Handler, StreamHandler
from logging.handlers import TimedRotatingFileHandler
from pathlib import Path

from core.settings import settings


GLOBAL_LOGGER: logging.Logger | None = None


def create_file_handler(formatter: Formatter, log_level: logging._Level) -> Handler:
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


def create_stream_handler(formatter: Formatter, log_level: logging._Level) -> Handler:
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


def get_logger() -> logging.Logger:
    global GLOBAL_LOGGER  # pylint: disable=global-statement
    log_level = logging.DEBUG if settings.arch_stats_dev_mode else logging.INFO

    if GLOBAL_LOGGER is None:
        logging.basicConfig(force=True, level=logging.DEBUG)
        GLOBAL_LOGGER = logging.getLogger(__name__)

        GLOBAL_LOGGER.setLevel(log_level)
        formatter = Formatter(
            "%(asctime)s|%(pathname)s:%(lineno)d|%(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
        )
        file_handler = create_file_handler(formatter, log_level)
        console_handler = create_stream_handler(formatter, log_level)
        GLOBAL_LOGGER.addHandler(file_handler)
        GLOBAL_LOGGER.addHandler(console_handler)
    return GLOBAL_LOGGER
