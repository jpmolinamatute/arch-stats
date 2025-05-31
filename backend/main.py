#!/usr/bin/env python

import logging
import asyncio
import argparse
from collections.abc import Callable, Awaitable

from shared import get_logger, LogLevel
from server.app import run as server_run
from target_reader.main import run as target_reader_run
from bow_reader.main import run as bow_reader_run
from arrow_reader.main import run as arrow_reader_run


def parse_args() -> argparse.Namespace:
    """
    Parse command-line arguments.
    """
    parser = argparse.ArgumentParser(
        description="Arch Stats main entry point. Start a module by name."
    )
    parser.add_argument(
        "module",
        type=str,
        choices=["server", "target_reader", "bow_reader", "arrow_reader"],
        help="Module to run.",
    )
    parser.add_argument(
        "--log-level",
        type=str,
        default="INFO",
        choices=[level.name for level in LogLevel],
        help="Logging level (default: INFO).",
    )
    return parser.parse_args()


def get_run_fn(module: str) -> Callable[[], Awaitable[None]]:
    """
    Import and return the async run function for the specified module.
    """
    runnable: Callable[[], Awaitable[None]]
    if module == "server":
        runnable = server_run
    elif module == "target_reader":
        runnable = target_reader_run
    elif module == "bow_reader":
        runnable = bow_reader_run
    elif module == "arrow_reader":
        runnable = arrow_reader_run
    else:
        raise ValueError(f"Unknown module: {module}")
    return runnable


async def run_with_shutdown(run_fn: Callable[[], Awaitable[None]], logger: logging.Logger) -> None:
    """
    Run the specified module, handling shutdown and exceptions gracefully.
    """
    try:
        await run_fn()
    except (KeyboardInterrupt, SystemExit):
        logger.info("Shutting down gracefully...")
    except Exception as e:
        logger.exception(f"Unhandled exception: {e!r}")
    finally:
        logger.info("Bye!")


def main() -> None:
    """
    Main entry point for the script.
    """
    args = parse_args()
    logger = get_logger(__name__, args.log_level)
    logger.info("Starting module: %s (log level: %s)", args.module, args.log_level)
    run_fn = get_run_fn(args.module)
    asyncio.run(run_with_shutdown(run_fn, logger))


if __name__ == "__main__":
    main()
