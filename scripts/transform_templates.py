# /// script
# requires-python = ">=3.14"
# dependencies = ["jinja2"]
# ///

import argparse
import sys
from pathlib import Path

from jinja2 import Environment, FileSystemLoader
from jinja2.exceptions import TemplateNotFound, TemplateSyntaxError


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Render Jinja2 templates using command line arguments."
    )
    parser.add_argument("--app-name", required=True, help="Name of the application")
    parser.add_argument("--app-name-label", required=True, help="Label of the application")
    parser.add_argument(
        "--prod-uvicorn-port",
        type=int,
        required=True,
        help="Production Uvicorn port number",
    )
    parser.add_argument(
        "--app-user-home-dir", required=True, help="Home directory of the application user"
    )
    parser.add_argument(
        "--cloudflared-tunnel-id", required=True, help="Cloudflared Tunnel ID"
    )
    parser.add_argument(
        "--output-dir", required=True, type=Path, help="Directory to save the transformed templates"
    )

    args = parser.parse_args()

    # Determine templates directory
    script_dir = Path(__file__).parent.resolve()
    templates_dir = script_dir / "templates"
    output_dir = args.output_dir

    if not templates_dir.is_dir():
        print(f"Error: Templates directory not found at {templates_dir}", file=sys.stderr)
        sys.exit(1)

    if not output_dir.is_dir():
        print(f"Error: Output directory not found at {output_dir}", file=sys.stderr)
        sys.exit(1)

    # Initialize Jinja2 Environment
    env = Environment(loader=FileSystemLoader(templates_dir), autoescape=False)
    context = {
        "app_name": args.app_name,
        "app_name_label": args.app_name_label,
        "prod_uvicorn_port": args.prod_uvicorn_port,
        "cloudflared_tunnel_id": args.cloudflared_tunnel_id,
        "app_user_home_dir": args.app_user_home_dir,
    }

    try:
        for template_name in ["arch-stats.service", "cloudflared_config.yaml"]:
            template = env.get_template(f"{template_name}.j2")
            rendered = template.render(context)
            dest_path = output_dir / template_name
            dest_path.write_text(rendered, encoding="utf-8")
            print(f"Successfully rendered and saved: {dest_path}")
    except TemplateNotFound:
        print(f"Error: Template {template_name} not found.", file=sys.stderr)
        sys.exit(1)
    except TemplateSyntaxError:
        print(f"Error: Template {template_name} has a syntax error.", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        print(f"Error rendering {template_name}: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
