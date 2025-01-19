#!/usr/bin/env python
import argparse
import asyncio
import sys
from os import getenv

import asyncpg
import uvicorn

from hub import setup as setup_hub
from server import create_app
from shared import get_logger
from shoot_recorder import setup as setup_shoot_recorder


def main() -> None:
    exit_status = 0
    logger = get_logger()
    parser = argparse.ArgumentParser()
    parser.add_argument("cmd", help="Command to run", choices=["server", "shoot_recorder", "hub"])
    args = parser.parse_args()
    server_name = getenv("ARCH_STATS_HOSTNAME", "localhost")
    server_port = int(getenv("ARCH_STATS_SERVER_PORT", "8000"))
    try:
        if args.cmd == "shoot_recorder":
            asyncio.run(setup_shoot_recorder(logger))
        elif args.cmd == "server":
            logger.info("Starting the server on %s:%d", server_name, server_port)
            app = create_app()
            uvicorn.run(app, host=server_name, port=server_port)
        elif args.cmd == "hub":
            asyncio.run(setup_hub(logger))
        else:
            logger.error("Unknown command %s", args.cmd)
            exit_status = 1
    except asyncpg.PostgresError as e:
        logger.exception("Database error occurred: %s", e)
        exit_status = 1
    except KeyboardInterrupt:
        logger.info("Process interrupted by user")
    except Exception as e:  # pylint: disable=broad-except
        logger.exception("An unexpected error occurred: %s", e)
        exit_status = 1
    finally:
        logger.info("Exiting with status %d", exit_status)
        sys.exit(exit_status)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("cmd", help="Command to run", choices=["server", "shoot_recorder", "hub"])
    args = parser.parse_args()

    asyncio.run(run_async(args.cmd))


if __name__ == "__main__":
    main()
