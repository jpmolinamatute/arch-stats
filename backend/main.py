#!/usr/bin/env python
import argparse
import asyncio
import signal
import sys
from os import getenv

import asyncpg
import uvicorn

from hub import setup as setup_hub
from server import create_app
from shared import get_logger
from shoot_recorder import setup as setup_shoot_recorder


logger = get_logger()


def set_async_handler() -> None:
    loop = asyncio.get_running_loop()
    stop_event = asyncio.Event()

    def handle_signal() -> None:
        stop_event.set()

    loop.add_signal_handler(signal.SIGTERM, handle_signal)
    loop.add_signal_handler(signal.SIGINT, handle_signal)


async def run_async(cmd: str) -> None:
    exit_status = 0

    try:
        set_async_handler()
        if cmd == "shoot_recorder":
            await setup_shoot_recorder(logger)
        elif cmd == "server":
            server_name = getenv("ARCH_STATS_HOSTNAME", "localhost")
            server_port = int(getenv("ARCH_STATS_SERVER_PORT", "8000"))
            logger.info("Starting the server on %s:%d", server_name, server_port)
            app = create_app()
            config = uvicorn.Config(app, host=server_name, port=server_port)
            server = uvicorn.Server(config)
            await server.serve()
        elif cmd == "hub":
            await setup_hub(logger)
        else:
            logger.error("Unknown command %s", cmd)
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
