import glob
import shutil
from pathlib import Path

from Cython.Build import cythonize
from invoke import Collection, Context, task
from setuptools import setup

from app.main import main
from docs import apps_communication, data_flow


PROJECT_ROOT = Path(__file__).parent
PYPROJECT = PROJECT_ROOT.joinpath("pyproject.toml")
APP_ROOT = PROJECT_ROOT.joinpath("app/")
DOCS_ROOT = PROJECT_ROOT.joinpath("docs/")
PTY = True
ECHO = True

ns = Collection()
lint = Collection("lint")
tests = Collection("tests")
app = Collection("app")
diagrams = Collection("diagrams")
convert_c = Collection("convert_c")


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


@task(name="compile_default")
def compile_default(_: Context) -> None:
    _log_open("Compiling default App")
    py_path = APP_ROOT.joinpath("/**/*.py")
    py_files = glob.glob(str(py_path), recursive=True)
    setup(ext_modules=cythonize(py_files, nthreads=4))


convert_c.add_task(compile_default)


@task(name="d_data_flow")
def d_data_flow(_: Context) -> None:
    _log_open("Running data flow diagrams")
    data_flow(DOCS_ROOT)


@task(name="d_apps_communication")
def d_apps_communication(_: Context) -> None:
    _log_open("Running apps communication diagrams")
    apps_communication(DOCS_ROOT)


diagrams.add_task(d_data_flow)
diagrams.add_task(d_apps_communication)


@task(name="default")
def run_app(_: Context) -> None:
    _log_open("Running default App")
    main()


app.add_task(run_app)


def _run_pylint(ctx: Context, ignore_failures: bool = True) -> None:
    cmd1 = f"pylint --rcfile {PYPROJECT} {APP_ROOT} {DOCS_ROOT}"
    _log_open("pylint")
    ctx.run(cmd1, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_black(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"black --config {PYPROJECT} {APP_ROOT} {DOCS_ROOT}"
    _log_open("black")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_isort(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"isort --settings-path {PYPROJECT} {APP_ROOT} {DOCS_ROOT}"
    _log_open("isort")
    ctx.run(cmd, pty=PTY, echo=ECHO, warn=ignore_failures)


def _run_mypy(ctx: Context, ignore_failures: bool = True) -> None:
    cmd = f"mypy --config-file {PYPROJECT} {APP_ROOT} {DOCS_ROOT}"
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
ns.add_collection(diagrams)
ns.add_collection(convert_c)
