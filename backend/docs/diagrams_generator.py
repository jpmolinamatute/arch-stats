#!/usr/bin/env python

from typing import Any, Mapping
from pathlib import Path
from graphviz import Digraph


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


def tournament_creation(**opt: Mapping[str, Any]) -> None:
    dot = Digraph(
        comment="Tournament creation flow.",
        graph_attr={
            "label": "Tournament creation flow.",
            "labelloc": "t",
            "labeljust": "c",
            "fontsize": "20",
        },
    )
    dot.node("creation_starts", shape="circle", label="Start", color="blue")
    dot.node("is_there_tournament", shape="diamond", label="Is there a tournament?")
    dot.node("is_there_tournament_yes", label="Yes.")
    dot.node("is_there_tournament_no", label="No.")
    dot.node("create_tournament", shape="box", label="Create tournament")
    dot.node("setup_lane", shape="box", label="Setup lane")
    dot.node("setup_target", shape="box", label="Setup target")
    dot.node("does_archer_exist", shape="diamond", label="Does archer exist?")
    dot.node("does_archer_exist_yes", label="Yes.")
    dot.node("does_archer_exist_no", label="No.")
    dot.node("is_archer_registered", shape="diamond", label="Is archer registered?")
    dot.node("is_archer_registered_yes", label="Yes.")
    dot.node("is_archer_registered_no", label="No.")
    dot.node("register_archer", shape="box", label="Register archer")
    dot.node("is_there_more_archers", shape="diamond", label="Is there more archers?")
    dot.node("is_there_more_archers_yes", label="Yes.")
    dot.node("is_there_more_archers_no", label="No.")
    dot.node("create_archer", shape="box", label="Create archer")
    dot.node("collect_archer_info", shape="box", label="Collect archer info")
    dot.node("collect_arrow_info", shape="box", label="Collect arrow info")
    dot.node("creation_ends", shape="doublecircle", label="End", color="red")

    dot.edge("creation_starts", "is_there_tournament")
    dot.edges(
        [
            ("is_there_tournament", "is_there_tournament_yes"),
            ("is_there_tournament", "is_there_tournament_no"),
        ]
    )
    dot.edge("is_there_tournament_yes", "does_archer_exist")
    dot.edge("is_there_tournament_no", "create_tournament")
    dot.edge("create_tournament", "setup_lane")
    dot.edge("setup_lane", "setup_target")
    dot.edge("setup_target", "does_archer_exist")
    dot.edges(
        [
            ("does_archer_exist", "does_archer_exist_yes"),
            ("does_archer_exist", "does_archer_exist_no"),
        ]
    )
    dot.edge("does_archer_exist_yes", "is_archer_registered")
    dot.edge("does_archer_exist_no", "create_archer")
    dot.edges(
        [
            ("is_archer_registered", "is_archer_registered_yes"),
            ("is_archer_registered", "is_archer_registered_no"),
        ]
    )
    dot.edge("is_archer_registered_no", "register_archer")
    dot.edge("create_archer", "collect_archer_info")
    dot.edge("collect_archer_info", "collect_arrow_info")
    dot.edge("collect_arrow_info", "register_archer")
    dot.edge("register_archer", "is_there_more_archers")
    dot.edges(
        [
            ("is_there_more_archers", "is_there_more_archers_yes"),
            ("is_there_more_archers", "is_there_more_archers_no"),
        ]
    )
    dot.edge("is_there_more_archers_yes", "setup_lane")
    dot.edge("is_there_more_archers_no", "creation_ends")
    dot.edge("is_archer_registered_yes", "creation_ends")
    dot.render("tournament_creation", **opt)


def main() -> None:
    docs_path = Path(__file__).parent
    opt: Mapping[str, Any] = {
        "overwrite_source": True,
        "format": "png",
        "view": True,
        "cleanup": True,
        "directory": docs_path,
    }

    tournament_creation(**opt)


if __name__ == "__main__":
    main()
