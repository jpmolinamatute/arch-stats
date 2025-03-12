import logging
import logging.handlers
import sys
from pathlib import Path


def get_logger(file_name: str = __name__) -> logging.Logger:
    logger = logging.getLogger(file_name)
    if not logger.hasHandlers():
        log_dir = Path(__file__).parent.joinpath("logs/")
        log_dir.mkdir(exist_ok=True)
        log_file = log_dir.joinpath("app.log")
        logger.setLevel(logging.INFO)

        # Add StreamHandler to output logs to console
        handler = logging.StreamHandler(sys.stdout)
        handler.setLevel(logging.INFO)
        formatter = logging.Formatter(
            "%(asctime)s|%(pathname)s:%(lineno)d|%(message)s", datefmt="%Y-%m-%d %H:%M:%S"
        )
        # Create a file handler
        file_handler = logging.FileHandler(log_file)
        file_handler.setLevel(logging.DEBUG)
        file_handler.setFormatter(formatter)

        # Create a console handler
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(logging.DEBUG)
        console_handler.setFormatter(formatter)

        # Add handlers to the logger
        logger.addHandler(file_handler)
        logger.addHandler(console_handler)
    return logger
