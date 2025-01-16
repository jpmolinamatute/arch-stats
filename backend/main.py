from math import e
from os import getenv
import sys
import argparse

import psycopg
import uvicorn

from server import create_app
from shared import get_logger
from shoot_recorder import setup


def main() -> None:
    exit_status = 0
    logger = get_logger()
    logger.info("Starting the hub...")
    parser = argparse.ArgumentParser()
    parser.add_argument("cmd", help="Command to run", choices=["server", "setup"])
    args = parser.parse_args()
    server_name = getenv("ARCH_STATS_HOSTNAME", "localhost")
    server_port = int(getenv("ARCH_STATS_SERVER_PORT", "8000"))
    try:
        if args.cmd == "setup":
            setup(logger)
        elif args.cmd == "server":
            app = create_app()
            uvicorn.run(app, host=server_name, port=server_port)
        else:
            logger.error("Unknown command %s", args.cmd)
            exit_status = 1
    except psycopg.Error:
        logger.exception("ERROR: a database error occurred")
        exit_status = 1
    except KeyboardInterrupt:
        logger.info("Bye!")
    except Exception:  # pylint: disable=broad-except
        logger.exception("An unexpected error occurred")
        exit_status = 1

    sys.exit(exit_status)


if __name__ == "__main__":
    main()
