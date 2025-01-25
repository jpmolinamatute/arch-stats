#!/usr/bin/env python
import argparse
import sys
from asyncio import run

import asyncpg

from hub import setup as setup_hub
from server import setup as setup_server
from shared import get_logger
from shoot_recorder import setup as setup_shoot_recorder


logger = get_logger()


async def run_async(cmd: str) -> None:
    if cmd == "shoot_recorder":
        await setup_shoot_recorder(logger)
    elif cmd == "server":
        await setup_server(logger)
    elif cmd == "hub":
        await setup_hub(logger)
    else:
        raise ValueError(f"Unknown command: {cmd}")


def main(module_name: str) -> None:
    exit_status = 0
    try:
        run(run_async(module_name))
    except asyncpg.PostgresError as e:
        logger.exception("Database error occurred: %s", e)
        exit_status = 1
    except KeyboardInterrupt:
        logger.info("Main process interrupted by user")
    except Exception as e:  # pylint: disable=broad-except
        logger.exception("An unexpected error occurred: %s", e)
        exit_status = 1
    finally:
        logger.info("Exiting with status %d", exit_status)
        sys.exit(exit_status)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("cmd", help="Command to run", choices=["server", "shoot_recorder", "hub"])
    args = parser.parse_args()
    main(args.cmd)
