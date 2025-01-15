import shutil
from pathlib import Path

from invoke.collection import Collection
from invoke.context import Context
from invoke.tasks import task


CURRENT_SCRIPT = Path(__file__).resolve()
PROJECT_ROOT = CURRENT_SCRIPT.parent
SERVER_ROOT = PROJECT_ROOT.joinpath("server")
HUB_ROOT = PROJECT_ROOT.joinpath("hub")
SHOOT_RECORDER_ROOT = PROJECT_ROOT.joinpath("shoot_recorder")
PYPROJECT = PROJECT_ROOT.joinpath("pyproject.toml")
MAIN = PROJECT_ROOT.joinpath("main.py")

PTY = True
ECHO = True


ns = Collection()
compose = Collection("compose")
lint = Collection("lint")
application = Collection("app")


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


def _run_pylint(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"pylint --rcfile {PYPROJECT} "
    cmd += f"{SERVER_ROOT} {HUB_ROOT} {SHOOT_RECORDER_ROOT} {CURRENT_SCRIPT} {MAIN}"
    _log_open("pylint")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_black(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"black --config {PYPROJECT} "
    cmd += f"{SERVER_ROOT} {HUB_ROOT} {SHOOT_RECORDER_ROOT} {CURRENT_SCRIPT}  {MAIN}"
    _log_open("black")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_isort(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"isort --settings-path {PYPROJECT} "
    cmd += f"{SERVER_ROOT} {HUB_ROOT} {SHOOT_RECORDER_ROOT} {CURRENT_SCRIPT}  {MAIN}"
    _log_open("isort")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_mypy(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"mypy --config-file {PYPROJECT} "
    cmd += f"{SERVER_ROOT} {HUB_ROOT} {SHOOT_RECORDER_ROOT} {CURRENT_SCRIPT}  {MAIN}"
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

ns.add_collection(lint)
