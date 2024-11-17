from pathlib import Path
import shutil

from invoke import task, Context, Collection

from app.src.main import main

CURRENT_SCRIPT = Path(__file__).resolve()
PROJECT_ROOT = CURRENT_SCRIPT.parent
PYPROJECT = PROJECT_ROOT.joinpath("/pyproject.toml")
APP_ROOT = PROJECT_ROOT.joinpath("/src/app")
PTY = True
ECHO = True

ns = Collection()
lint = Collection("lint")
tests = Collection("tests")
app = Collection("app")


def _log_open(msg: str) -> None:
    terminal_size = shutil.get_terminal_size((80, 20))
    char_to_use = "="
    width = terminal_size.columns
    print(f"\n{char_to_use*width}")
    padding = (width - len(msg) - 12) // 2
    msg = f"{char_to_use*padding} Running '{msg}' {char_to_use*padding}"
    if len(msg) < width:
        msg += char_to_use
    print(msg)
    print(f"{char_to_use*width}\n")


@task(name="default")
def run_app(_: Context) -> None:
    _log_open("Running default App")
    main()


app.add_task(run_app)


def _run_pylint(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"pylint --rcfile {PYPROJECT} {APP_ROOT}"
    _log_open("pylint")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_black(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"black --config {PYPROJECT} {APP_ROOT}"
    _log_open("black")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_isort(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"isort --settings-path {PYPROJECT} {APP_ROOT}"
    _log_open("isort")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_mypy(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"mypy --config-file {PYPROJECT} {APP_ROOT}"
    _log_open("mypy")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


@task(name="pylint")
def pylint(ctx: Context) -> None:
    _run_pylint(ctx)


@task(name="black")
def black(ctx: Context) -> None:
    _run_black(ctx)


@task(name="isort")
def isort(ctx: Context) -> None:
    _run_isort(ctx)


@task(name="mypy")
def mypy(ctx: Context) -> None:
    _run_mypy(ctx)


@task(name="run_all")
def run_all(ctx: Context, ignore_failures: bool = True) -> None:
    print("Running ALL linting tools")
    _run_black(ctx, ignore_failures)
    _run_isort(ctx, ignore_failures)
    _run_mypy(ctx, ignore_failures)
    _run_pylint(ctx, ignore_failures)


lint.add_task(pylint)
lint.add_task(black)
lint.add_task(isort)
lint.add_task(mypy)
lint.add_task(run_all)


@task(name="pytest")
def pytest(ctx: Context) -> None:
    test_path = APP_ROOT.joinpath("/src/tests")
    cmd = f"pytest --config-file={PYPROJECT} {test_path}"
    _log_open("pytest")
    ctx.run(cmd, pty=PTY, echo=ECHO)


tests.add_task(pytest)
ns.add_collection(lint)
ns.add_collection(tests)
ns.add_collection(app)
