import argparse
import json
import os
import sys
from pathlib import Path

from app import run


def generate_openapi_spec(output_path: Path) -> None:
    app = run()
    schema = app.openapi()
    print(f"Writing OpenAPI schema to: {output_path}")
    output_path.write_text(json.dumps(schema, indent=4))


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate OpenAPI JSON schema.")
    parser.add_argument("output_path", type=Path, help="Path to save the OpenAPI JSON file")
    args = parser.parse_args()

    output_dir = args.output_path.parent

    exit_code = 0
    try:
        if not output_dir.is_dir():
            raise NotADirectoryError(f"'{output_dir}' is not a directory.")
        if not os.access(output_dir, os.W_OK):
            raise PermissionError(f"Directory '{output_dir}' is not writable.")

        generate_openapi_spec(args.output_path)
        print("Script completed successfully")
    except Exception as e:
        print(f"An error occurred: {e}", file=sys.stderr)
        exit_code = 1
    finally:
        sys.exit(exit_code)


if __name__ == "__main__":
    main()
