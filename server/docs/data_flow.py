from pathlib import Path

from diagrams import Diagram
from diagrams.custom import Custom
from diagrams.programming.framework import Angular, Fastapi
from diagrams.programming.language import Rust

# from diagrams.generic.os import Raspbian
# raspberry_pi = Raspbian("Raspberry Pi 5")


def data_flow(docs_root: Path) -> None:
    with Diagram(name="Data flow", filename="arch_stats_data_flow", show=False):
        # pylint: disable=expression-not-assigned,pointless-statement
        image_path = docs_root.joinpath("img/postgresql.png")
        postgresql = Custom(label="PostgreSQL", icon_path=image_path)
        server = Fastapi("FastAPI")
        webui = Angular("Angular")
        shot_reader = Rust("Shot Reader")
        arrow_reader = Rust("Arrow Reader")
        bow_reader = Rust("Bow Reader")

        [shot_reader, arrow_reader, bow_reader] >> postgresql >> server >> webui
