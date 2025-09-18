import sys
from pathlib import Path
import json

from app import run


def generate_openapi_spec() -> None:
    repo_root = Path(__file__).resolve().parent.parent
    app = run()
    schema = app.openapi()
    output_path = repo_root / "openapi.json"
    print(output_path)
    output_path.write_text(json.dumps(schema, indent=4))


def main() -> None:
    exit_code = 0
    try:
        generate_openapi_spec()
    except Exception as e:
        print(f"An error occurred: {e}", file=sys.stderr)
        exit_code = 1
    finally:
        if exit_code != 0:
            print("Exiting with error code", exit_code, file=sys.stderr)
        else:
            print("Script completed successfully")
    sys.exit(exit_code)


if __name__ == "__main__":
    main()
