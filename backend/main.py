#!/usr/bin/env python


import argparse

import sys
from asyncio import run

import asyncpg
from dotenv import load_dotenv

# from hub import setup as setup_hub
from server import setup as setup_server
from shared import get_logger, LogLevel
from target_reader import setup as setup_target_reader


logger = get_logger(__name__, LogLevel.INFO)


async def run_async(cmd: str) -> None:
    if cmd == "target_reader":
        await setup_target_reader(logger)
    elif cmd == "server":
        await setup_server(logger)
    elif cmd == "hub":
        pass
    else:
        raise ValueError(f"Unknown command: {cmd}")


def main(module_name: str) -> None:
    exit_status = 0
    try:
        load_dotenv()
        run(run_async(module_name))
    except asyncpg.PostgresError as e:
        logger.exception("Database error occurred: %s", e)
        exit_status = 1
    except KeyboardInterrupt:
        logger.info("Main process interrupted by user")
    except Exception as e:
        logger.exception("An unexpected error occurred: %s", e)
        exit_status = 1
    finally:
        logger.info("Exiting with status %d", exit_status)
        sys.exit(exit_status)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("cmd", help="Command to run", choices=["server", "target_reader", "hub"])
    args = parser.parse_args()
    main(args.cmd)
