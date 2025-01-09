import sys


def main() -> None:
    print("Hello from arch-stats!")


if __name__ == "__main__":
    EXIT_STATUS = 0
    try:
        main()
    except Exception as e:  # pylint: disable=broad-exception-caught
        print(e)
        EXIT_STATUS = 1
    finally:
        sys.exit(EXIT_STATUS)
