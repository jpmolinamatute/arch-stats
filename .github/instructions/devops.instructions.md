# DevOps instructions

**Audience:** DevOps developers working in [script/](../../scripts/), and [.github/workflows](../workflows)

---
applyTo: "**/*.bash,.github/workflows/*.yaml"
---


**Goal:** When suggesting or completing any `*.bash` or shell snippet, Copilot must produce **safe, readable, and lintable** Bash that passes `shellcheck` and formats with `shfmt`.

## Language & Environment

* Use **Bash** (not POSIX sh). Assume **Linux**.
* Always start scripts with this shebang (no extra spaces before/after `#!`):

```sh
#!/usr/bin/env bash
```

* Use **4-space indentation** to match the repo style.

## Strict Mode & Safety Defaults

Copilot must include these at the top of scripts, beneath the shebang:

```sh
set -Eeuo pipefail
```

* `set -e`: exit on error
* `set -u`: undefined var is an error
* `set -o pipefail`: fail pipelines early
* `set -E`: preserve ERR traps in functions

Prefer **double quotes** around expansions, use **`local`** inside functions, and avoid word-splitting surprises.

## Canonical Script Entry Point

Every standalone script must provide a `main` function and call it with the original args. Exit explicitly with `0` on success.

```sh
main() {
    echo "Hello World"
    exit 0
}

main "$@"
```

> Note: The call must be `main "$@"` (no stray braces).

## Error Messages & Exit Codes

When generating error paths, **write to stderr** and **exit non-zero**:

```sh
echo "ERROR: something went wrong!" >&2
exit 2
```

Prefer distinct codes per failure class where useful (e.g., `2` for usage, `3` for missing file, `4` for network, etc.).

## Minimal Logging Helpers (optional but encouraged)

```sh
log_info()  { echo "INFO: $*"; }
log_warn()  { echo "WARN: $*" >&2; }
log_error() { echo "ERROR: $*" >&2; }
```

Use these consistently; all errors must go to stderr.

## Argument Parsing Pattern

When scripts take flags, use `getopts` and show a usage on error:

```sh
usage() {
    cat >&2 <<'EOF'
Usage: script.bash [-f FILE] [-n NUM]
  -f FILE   input file (required)
  -n NUM    sample size (default: 10)
EOF
    exit 2
}

main() {
    local file=""
    local num="10"

    while getopts ":f:n:h" opt; do
        case "$opt" in
            f) file="$OPTARG" ;;
            n) num="$OPTARG" ;;
            h) usage ;;
            \?) echo "ERROR: invalid option: -$OPTARG" >&2; usage ;;
            :)  echo "ERROR: option -$OPTARG requires an argument" >&2; usage ;;
        esac
    done
    shift $((OPTIND - 1))

    [[ -n "$file" ]] || { echo "ERROR: -f FILE is required" >&2; exit 2; }
    [[ -r "$file" ]] || { echo "ERROR: cannot read file: $file" >&2; exit 3; }

    # script logic...
    exit 0
}

main "$@"
```

## Filesystem & Process Checks

* Before running commands that depend on files/directories, check first:

   ```sh
   [[ -f "$path" ]] || { echo "ERROR: file not found: $path" >&2; exit 3; }
   [[ -d "$dir"  ]] || { echo "ERROR: directory not found: $dir" >&2; exit 3; }
   ```

* For arrays and globs, use `"${arr[@]}"` and `shopt -s nullglob` when needed (avoid SC2048/SC2086 issues).

## shellcheck Rules & In-File Directives

Copilot should write code that **passes `shellcheck` without suppressions**. If a directive is unavoidable, make it **narrow and justified**:

```sh
# shellcheck disable=SC2312 # Justification: we intentionally need $? from the pipeline's last command here.
```

Frequent pitfalls to avoid:

* **SC2046/SC2086**: Always quote expansions unless array spreading is intended.
* **SC2164**: Check `cd` success: `cd "$dir" || { echo "ERROR: cd $dir" >&2; exit 4; }`
* **SC2155**: Prefer separate `local` and assignment if it improves clarity.
* **SC2181**: Prefer `if cmd; then ...` over checking `$?` directly.

## Formatting with shfmt

Copilot should produce code that formats predictably with `shfmt`. Target options:

* Indentation: **4 spaces**
* Binary ops at line end
* Case indentation
* Simplify redirects

Equivalent CLI (for humans):

```bash
shfmt -i 4 -bn -ci -sr -w .
```

> Copilot: write code that remains stable under these rules (e.g., braces/`do`/`done` on their own lines).

## Subprocess & Pipe Guidelines

* Prefer explicit pipelines:

  ```sh
  if ! curl -fsS "$url" -o "$out"; then
      echo "ERROR: download failed: $url" >&2
      exit 5
  fi
  ```

* For `grep`/`awk`/`sed`, quote patterns and variables; avoid useless `cat`.
* Use `printf` for formatted output; `echo -e` is discouraged.

## Traps & Cleanup (when creating temp resources)

If generating temp files or background jobs, add a cleanup trap:

```sh
tmp_dir="$(mktemp -d)"
cleanup() { rm -rf "$tmp_dir"; }
trap cleanup EXIT INT TERM
```

## Arrays & SCP/RSYNC Patterns

When suggesting multi-file operations, **verify each file** before transfer and expand arrays safely:

```sh
upload_scripts() {
    local host="$1"; shift
    local files=("$@")

    ((${#files[@]})) || { echo "ERROR: no files provided" >&2; exit 2; }

    local f
    for f in "${files[@]}"; do
        [[ -f "$f" ]] || { echo "ERROR: missing file: $f" >&2; exit 3; }
    done

    scp -o BatchMode=yes -o ConnectTimeout=5 -- "${files[@]}" "${host}:/tmp/" \
        || { echo "ERROR: scp failed to ${host}" >&2; exit 4; }
}
```

> Note the `--` before `"${files[@]}"` to guard against files beginning with `-`.

## Network Timeouts & Retries

* Always set **timeouts** on network calls (`curl -m 10 --retry 2 --retry-all-errors`).
* Validate success and fail loudly to stderr.

## Environment Variables & Configuration Injection

Deployment / CI jobs must surface required runtime variables explicitly (mirror backend instructions):

| Variable | Purpose | Required | Default (dev) |
|----------|---------|----------|---------------|
| `PGHOST` | Postgres host | yes | `localhost` |
| `PGPORT` | Postgres port | no | `5432` |
| `PGUSER` | Postgres user | yes | `postgres` |
| `PGPASSWORD` | Postgres password | yes | *(secret)* |
| `PGDATABASE` | DB name | yes | `arch_stats` |
| `PG_POOL_MIN` | Min pool connections | no | `1` |
| `PG_POOL_MAX` | Max pool connections | no | `10` |
| `API_MAX_CONCURRENT_REQUESTS` | Future semaphore limit | no | `100` |

Guidelines:
* Provide `.env.example` (non-secret placeholders) when adding new variables.
* CI workflows should mask secrets (`***`) and never echo raw secret values.
* Validate required variables early in startup scripts with a helper:

```bash
require_var() { local n="$1"; [[ -n "${!n:-}" ]] || { echo "ERROR: required env var $n not set" >&2; exit 2; }; }
for v in PGHOST PGUSER PGPASSWORD PGDATABASE; do require_var "$v"; done
```

## Migrations Alignment

DevOps workflows running migrations must:
1. Execute pending SQL files (lexicographic order) before app startup.
2. Fail fast if a migration partially applies (wrap each file in a transaction where safe).
3. Surface clear logs: `APPLY migration <file> ... OK` / `SKIP already applied`.
4. Record applied filenames in a bookkeeping table `schema_migrations (filename text primary key, applied_at timestamptz not null default now())` if/when automation is added.

Until an automated runner exists, document manual application steps in release notes.

## What Copilot Should **Always** Do

* Insert the **shebang**, **strict mode**, and **main** pattern.
* Write **quoted**, **lint-clean**, **4-space indented** Bash.
* Emit **stderr** on errors and **non-zero** exits.
* Offer **usage** when parsing args.
* Prefer **functions**, `local` vars, and early returns.

## What Copilot Should **Avoid**

* Unquoted expansions; naked `"$@"` is the exception.
* Silent failures; missing exit codes.
* Inline `shellcheck` suppressions without justification.
* `echo -e` for formatting; use `printf`.

