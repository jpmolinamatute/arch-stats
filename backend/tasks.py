import shutil
from os import getenv
from pathlib import Path
from time import sleep

from invoke.collection import Collection
from invoke.context import Context
from invoke.tasks import task

from main import main


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
test = Collection("test")


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


@task(name="wscat")
def run_wscat(ctx: Context) -> None:
    _log_open("Running wscat")
    hostname = getenv("ARCH_STATS_HOSTNAME", "localhost")
    port_number = getenv("ARCH_STATS_SERVER_PORT", "8000")
    ctx.run(
        f"wscat -c ws://{hostname}:{port_number}/ws",
        pty=PTY,
        echo=ECHO,
    )


test.add_task(run_wscat)


@task(name="hub")
def run_hub(_: Context) -> None:
    _log_open("Running hub")
    main("hub")


@task(name="server")
def run_server(_: Context) -> None:
    _log_open("Running server")
    main("server")


@task(name="shoot_recorder")
def run_shoot_recorder(_: Context) -> None:
    _log_open("Running shoot recorder")
    main("shoot_recorder")


application.add_task(run_hub)
application.add_task(run_server)
application.add_task(run_shoot_recorder)


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


def _seed(ctx: Context) -> None:
    _log_open("Seed database")
    project_parent = PROJECT_ROOT.parent
    seed_script = project_parent.joinpath("scripts/db-init.sh")
    ctx.run(str(seed_script), pty=PTY, echo=ECHO)


@task(name="up")
def compose_up(ctx: Context) -> None:
    _log_open("docker-compose up")
    project_parent = PROJECT_ROOT.parent
    docker_file = project_parent.joinpath("docker/docker-compose.yaml")
    ctx.run(f"docker compose -f {docker_file} up --detach", pty=PTY, echo=ECHO)
    sleep(2)
    _seed(ctx)


@task(name="down")
def compose_down(ctx: Context) -> None:
    _log_open("docker-compose down -v")
    project_parent = PROJECT_ROOT.parent
    docker_file = project_parent.joinpath("docker/docker-compose.yaml")
    ctx.run(f"docker compose -f {docker_file} down -v", pty=PTY, echo=ECHO)


@task(name="seed")
def compose_seed(ctx: Context) -> None:
    _seed(ctx)


compose.add_task(compose_up)
compose.add_task(compose_down)
compose.add_task(compose_seed)

ns.add_collection(compose)
ns.add_collection(lint)
ns.add_collection(application)
ns.add_collection(test)
