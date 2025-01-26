import logging
import sys


def get_logger(file_name: str = __name__) -> logging.Logger:
    logger = logging.getLogger(file_name)
    if not logger.hasHandlers():
        logger.setLevel(logging.INFO)

        # Add StreamHandler to output logs to console
        handler = logging.StreamHandler(sys.stdout)
        handler.setLevel(logging.INFO)
        formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
        handler.setFormatter(formatter)
        logger.addHandler(handler)
    return logger
