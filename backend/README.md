# Arch-Stats Backend

Welcome to the Arch-Stats backend! This guide will help you get up and running quickly, especially if you're new to the project or backend development.

## 🚀 What This Backend Does

Arch-Stats collects and manages archery performance data using Python 3.13. It's made up of:

* A **[FastAPI](https://fastapi.tiangolo.com/) [server](./src/server/)** that handles API requests and coordinates data.
* Three **sensor modules**:
  * **[Arrow Reader](./src/arrow_reader/)** – assigns unique IDs to arrows.
  * **[Bow Reader](./src/bow_reader/)** – logs when arrows are drawn/released.
  * **[Target Reader](./src/target_reader/)** – logs arrow impact location/time.

These 4 modules don't talk to each other directly. Instead, they all write to a **[PostgreSQL](https://www.postgresql.org/docs/) database** and use **[NOTIFY](https://www.postgresql.org/docs/current/sql-notify.html)/[LISTEN](https://www.postgresql.org/docs/current/sql-listen.html)** events to communicate. This decoupled setup makes everything modular and easy to test separately.

## 🧰 What You'll Need

Before you start, make sure you have:

* **uv** – Required. Used for managing both the virtual environment and dependencies. Install it from [uv docs](https://docs.astral.sh/uv).
* **Python 3.13+** – You must have Python 3.13 installed. We recommend installing it via `uv`.
* **Docker and Docker Compose** – Required to run PostgreSQL.
* **Linux Environment** – Native Linux or WSL recommended. (macOS/Windows may need extra setup.)
* **VS Code (optional)** – Project includes helpful workspace settings and tasks.

## 🛠️ Setup: Step-by-Step

### 1.  Install Dependencies

```bash
cd arch-stats/backend
uv sync --group dev --python $(cat ./.python-version)
```

This creates a `.venv` and installs everything listed in [pyproject.toml](./pyproject.toml).

### 2. Activate Your Environment

```bash
source .venv/bin/activate
```

Check Python version:

```bash
python --version  # should be +3.13
```

### 3. Start PostgreSQL (via Docker Compose)

Option 1: Using VS Code Tasks:

* Shortcut: `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux), then type and select `Tasks: Run Task`
* Choose **Start Docker Compose** to launch the database
* Choose **Stop Docker Compose** to shut it down when you're done

Option 2: From the terminal:

```bash
docker compose -f ./docker/docker-compose.yaml up
```

To stop:

```bash
docker compose -f ./docker/docker-compose.yaml down
```

### 4. Run the Dev Server & Bots

#### Run FastAPI Server

Option 1: Manually

```bash
cd backend/src
source ../.venv/bin/activate
uvicorn --loop uvloop --lifespan on --reload --ws websockets --http h11 --use-colors --log-level debug --timeout-graceful-shutdown 10 --factory --limit-concurrency 10 server.app:run
```

Option 2: Using VS Code Task

* Shortcut: `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux), then select `Tasks: Run Task`
* Choose **Start Uvicorn Dev Server**

#### Run Archy Bot (simulated target reader)

Option 1: Manually

```bash
cd backend/src
source ../.venv/bin/activate
./target_reader/archy.py
```

Option 2: Using VS Code Task

* Shortcut: `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux), then select `Tasks: Run Task`
* Choose **Start Archy bot**

Once running:

* API available at <http://localhost:8000>
* WebSocket stream updates in real-time from Archy

## 📁 Project Structure

```text
backend/
├── pyproject.toml     # central config
├── src/
│   ├── server/        # FastAPI server
│   ├── arrow_reader/  # arrow module
│   ├── bow_reader/    # bow module
│   ├── target_reader/ # target module
│   └── shared/        # utilities shared across modules
└── tests/             # all tests
```

## 🧹 Linting, Formatting & Testing

Use these tools to keep the codebase clean and reliable:

| Tool                                                      | Purpose                | Run It With                                    |
| --------------------------------------------------------- | ---------------------- | ---------------------------------------------- |
| [black](https://black.readthedocs.io/en/stable/)          | Code formatter         | `black --config ./pyproject.toml ./src`        |
| [isort](https://pycqa.github.io/isort/)                   | Sorts imports          | `isort --settings-file ./pyproject.toml ./src` |
| [mypy](https://mypy.readthedocs.io/en/stable/)            | Type checking (strict) | `mypy --config-file ./pyproject.toml ./src`    |
| [pylint](https://pylint.readthedocs.io/en/stable/)        | Linting for bugs/style | `pylint --rcfile ./pyproject.toml ./src`       |
| [pytest](https://docs.pytest.org/en/stable/contents.html) | Runs all tests         | `pytest -v`                                    |

To run everything at once:

```bash
./tools/linting.sh
```

> Linting and Auto-formatting are enabled if you use VS Code with [.vscode/settings.json](../.vscode/settings.json)

## 🧪 How Testing Works

* Tests live in [backend/tests/](./tests/)
* We use [pytest-asyncio](https://pytest-asyncio.readthedocs.io/en/stable/) to test `async def` code.
* Tests connect to the same PostgreSQL DB (launched by Docker Compose).
* FastAPI is tested using in-memory HTTP requests via `httpx.AsyncClient`
* You can use `faker` to generate dummy data if needed.

> **Pro Tip:** Make sure Docker (and the DB) is running before launching tests.

## ✅ You're Ready

Once you've completed the setup and verified the backend is running:

* Visit <http://localhost:8000/api/swagger> to view the live API docs.
* Open the WebUI to see real-time shot data.
* Hack away!
