# Template Rendering Script Design

Design for upgrading `scripts/transfor_templates.py` to collect parameters and render Jinja2 templates.

## Context
We have two Jinja2 templates under `scripts/templates/`:
1. `arch-stats.service.j2` (systemd unit template)
2. `config_template.yaml.j2` (cloudflared configuration template)

These templates require five variables:
- `app_name`
- `app_name_label`
- `prod_uvicorn_port`
- `cloudflared_tunnel_id`
- `app_user_home_dir`

## CLI Design
We will use standard `argparse` to parse:
- `--app-name` (str, required)
- `--app-name-label` (str, required)
- `--prod-uvicorn-port` (int, required)
- `--cloudflared-tunnel-id` (str, required)
- `--app-user-home-dir` (str, required)

## Dependency Management
We'll declare `jinja2` as a dependency in the inline script metadata block of `transfor_templates.py`:
```python
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "jinja2",
# ]
# ///
```
This enables seamless execution using `uv run scripts/transfor_templates.py`.

## Template Rendering
We will load templates using Jinja2's `FileSystemLoader` pointing to the template directory and render them with the parsed CLI argument namespace.
Rendered output will be written to stdout with demarcations, as the target files destinations will be discussed in a next step.
