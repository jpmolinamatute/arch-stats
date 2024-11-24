from pathlib import Path

from diagrams import Diagram
from diagrams.programming.framework import Angular, Fastapi
from diagrams.programming.language import Rust


def apps_communication(_: Path) -> None:
    with Diagram(name="Apps communication", filename="arch_stats_app_communication", show=False):
        # pylint: disable=expression-not-assigned,pointless-statement

        server = Fastapi("FastAPI")
        webui = Angular("Angular")
        shot_reader = Rust("Shot Reader")
        arrow_reader = Rust("Arrow Reader")
        bow_reader = Rust("Bow Reader")

        [shot_reader, arrow_reader, bow_reader, server] >> webui
