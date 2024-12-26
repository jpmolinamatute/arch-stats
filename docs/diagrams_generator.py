#!/usr/bin/env python

from typing import Any, Mapping
from pathlib import Path
from graphviz import Digraph


def lane_setup(**opt: Mapping[str, Any]) -> None:
    dot = Digraph(
        comment="Shooting flow.",
        node_attr={"color": "yellow"},
        edge_attr={"color": "yellow"},
        graph_attr={
            "label": "Shooting flow.",
            "labelloc": "t",
            "labeljust": "c",
            "fontsize": "20",
        },
    )
    dot.node(name="set_lane", label="Set a lane for the tournament.", shape="box")
    dot.node(name="get_target_hight", label="Measure target height", shape="box", style="rounded")
    dot.node(name="get_lane_length", label="Measure lane length", shape="box", style="rounded")
    dot.node(
        name="get_x_rings", label="Measure X coord\nfor each ring", shape="box", style="rounded"
    )
    dot.node("start", shape="circle", label="Start")
    dot.node("end", shape="doublecircle", label="End")
    dot.render("shooting_flow", **opt)


def shooting_flow(**opt: Mapping[str, Any]) -> None:
    dot = Digraph(
        comment="Shooting flow.",
        node_attr={"color": "yellow"},
        edge_attr={"color": "yellow"},
        graph_attr={
            "label": "Shooting flow.",
            "labelloc": "t",
            "labeljust": "c",
            "fontsize": "20",
        },
    )
    dot.node("start", shape="circle", label="Start")
    dot.node("end", shape="doublecircle", label="End")
    dot.render("shooting_flow", **opt)


def archer_tournament_creation(**opt: Mapping[str, Any]) -> None:
    dot = Digraph(
        comment="Archer creation flow.",
        node_attr={"color": "blue"},
        edge_attr={"color": "blue"},
        graph_attr={
            "label": "Archer creation flow.",
            "labelloc": "t",
            "labeljust": "c",
            "fontsize": "20",
        },
    )
    dot.node("archer_start", shape="circle", label="Start")
    dot.node("tournament_start", shape="circle", label="Start")
    dot.node(name="archer_exist", label="Does the archer exist?", shape="diamond")
    dot.node(name="exist_yes", label="Yes.")
    dot.node(name="archer_exist_no", label="No.")
    dot.node(name="ready", label="Ready to register\nin a tournament.", shape="box")
    dot.node(name="collect_info", label="Collect archer's info.", shape="box")
    dot.node(name="collect_arrow_info", label="Collect archer's\narrows info.", shape="box")
    dot.node(name="tournament_exist", label="Is the tournament created?", shape="diamond")
    dot.node(name="tournament_exist_no", label="No.")
    dot.node(name="create_tournament", label="Create a tournament", shape="box")
    dot.node(name="ready", label="Ready to register\nin a tournament.", shape="box")

    dot.edge("archer_start", "archer_exist")
    dot.edges([("archer_exist", "exist_yes"), ("archer_exist", "archer_exist_no")])
    dot.edge("archer_exist_no", "collect_info")
    dot.edges([("collect_info", "ready"), ("collect_info", "collect_arrow_info")])
    dot.edge("exist_yes", "ready")

    dot.edge("tournament_start", "tournament_exist")
    dot.edges(
        [
            ("tournament_exist", "exist_yes"),
            ("tournament_exist", "tournament_exist_no"),
        ]
    )
    dot.edge("tournament_exist_no", "create_tournament")

    dot.edge("create_tournament", "ready")

    dot.edge("ready", "end")
    dot.render("archer_creation", **opt)


def main() -> None:
    docs_path = Path(__file__).parent
    opt: Mapping[str, Any] = {
        "overwrite_source": True,
        "format": "png",
        "view": True,
        "cleanup": True,
        "directory": docs_path,
    }

    archer_tournament_creation(**opt)


if __name__ == "__main__":
    main()
