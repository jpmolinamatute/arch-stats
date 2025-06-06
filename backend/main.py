#!/usr/bin/env python

import logging
import asyncio
import argparse
from collections.abc import Callable, Awaitable

from shared import get_logger, LogLevel
from server import run as server_run
from target_reader import run as archy_run
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
        choices=["server", "target_reader", "bow_reader", "arrow_reader", "archy"],
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


def get_run_fn(module: str) -> Callable[[logging.Logger], Awaitable[None]]:
    """
    Import and return the async run function for the specified module.
    """
    runnable: Callable[[logging.Logger], Awaitable[None]]
    if module == "server":
        runnable = server_run
    elif module == "archy":
        runnable = archy_run
    elif module == "bow_reader":
        runnable = bow_reader_run
    elif module == "arrow_reader":
        runnable = arrow_reader_run
    else:
        raise ValueError(f"Unknown module: {module}")
    return runnable


async def run_with_shutdown(
    logger: logging.Logger, run_fn: Callable[[logging.Logger], Awaitable[None]]
) -> None:
    """
    Run the specified module, handling shutdown and exceptions gracefully.
    """
    try:
        await run_fn(logger)
    except (KeyboardInterrupt, SystemExit, asyncio.CancelledError):
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
    try:
        asyncio.run(run_with_shutdown(logger, run_fn))
    except KeyboardInterrupt:
        logger.info("KeyboardInterrupt received, shutting down...")


if __name__ == "__main__":
    main()
