#!/usr/bin/env python

import logging
import sys
from pathlib import Path


def main() -> None:
    exit_status = 0
    logging.basicConfig(level=logging.DEBUG)
    try:
        current_script = Path(__file__).resolve().name
        logging.info("Script %s has started", current_script)

        logging.info("Bye!")
    except Exception as err:
        logging.exception(err)
        exit_status = 1
    finally:
        sys.exit(exit_status)


if __name__ == "__main__":
    main()
