# Template Rendering Script Upgrade Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Update `scripts/transfor_templates.py` to collect all five parameters via command line flags and render the systemd and cloudflared templates.

**Architecture:** Use `argparse` to parse parameters, define dependencies in the script metadata block, and use Jinja2 to render the files to `stdout`.

**Tech Stack:** Python 3.12+, Jinja2, uv.

---

### Task 1: Update `scripts/transfor_templates.py`

**Files:**
- Modify: `scripts/transfor_templates.py`

**Step 1: Implement argument parsing and Jinja2 rendering**
Replace the content of `scripts/transfor_templates.py` with:
```python
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "jinja2",
# ]
# ///

import argparse
import sys
from pathlib import Path
from jinja2 import Environment, FileSystemLoader

def main() -> None:
    parser = argparse.ArgumentParser(description="Render Jinja2 templates using command line arguments.")
    parser.add_argument("--app-name", required=True, help="Name of the application")
    parser.add_argument("--app-name-label", required=True, help="Label of the application")
    parser.add_argument("--prod-uvicorn-port", type=int, required=True, help="Production Uvicorn port number")
    parser.add_argument("--cloudflared-tunnel-id", required=True, help="Cloudflared tunnel ID")
    parser.add_argument("--app-user-home-dir", required=True, help="Home directory of the application user")

    args = parser.parse_args()

    # Determine templates directory
    script_dir = Path(__file__).parent.resolve()
    templates_dir = script_dir / "templates"
    
    if not templates_dir.is_dir():
        print(f"Error: Templates directory not found at {templates_dir}", file=sys.stderr)
        sys.exit(1)

    # Initialize Jinja2 Environment
    env = Environment(
        loader=FileSystemLoader(templates_dir),
        autoescape=False
    )

    templates = [
        "arch-stats.service.j2",
        "config_template.yaml.j2"
    ]

    context = {
        "app_name": args.app_name,
        "app_name_label": args.app_name_label,
        "prod_uvicorn_port": args.prod_uvicorn_port,
        "cloudflared_tunnel_id": args.cloudflared_tunnel_id,
        "app_user_home_dir": args.app_user_home_dir,
    }

    for template_name in templates:
        try:
            template = env.get_template(template_name)
            rendered = template.render(context)
            
            print("=" * 80)
            print(f"RENDERED TEMPLATE: {template_name}")
            print("=" * 80)
            print(rendered)
            print("\n")
        except Exception as e:
            print(f"Error rendering {template_name}: {e}", file=sys.stderr)
            sys.exit(1)

if __name__ == "__main__":
    main()
```

**Step 2: Run the script to verify CLI and rendering output**

Run:
```bash
uv run scripts/transfor_templates.py \
  --app-name arch-stats \
  --app-name-label "Arch Stats" \
  --prod-uvicorn-port 8000 \
  --cloudflared-tunnel-id "some-tunnel-uuid" \
  --app-user-home-dir "/home/arch-stats"
```

Verify that the output contains both templates correctly rendered.

**Step 3: Run the script with missing arguments to verify errors**

Run:
```bash
uv run scripts/transfor_templates.py --app-name arch-stats
```

Expected output:
```
error: the following arguments are required: --app-name-label, --prod-uvicorn-port, --cloudflared-tunnel-id, --app-user-home-dir
```
