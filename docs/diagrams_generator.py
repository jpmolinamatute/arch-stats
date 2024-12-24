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
        node_attr={"color": "red"},
        edge_attr={"color": "red"},
        graph_attr={
            "label": "Tournament creation flow.",
            "labelloc": "t",
            "labeljust": "c",
            "fontsize": "20",
        },
    )
    dot.node("start", shape="circle", label="Start")
    dot.node(name="is_tournament_created", label="Is the tournament created?", shape="diamond")
    dot.node(name="is_tournament_created_yes", label="Yes.")
    dot.node(name="is_tournament_created_no", label="No.")
    dot.node(name="create_tournament", label="Create a tournament", shape="box")
    dot.node(name="set_lane", label="Set a lane for the tournament.", shape="box")
    dot.node(name="get_target_hight", label="Measure target height", shape="box", style="rounded")
    dot.node(name="get_lane_length", label="Measure lane length", shape="box", style="rounded")
    dot.node(
        name="get_x_rings", label="Measure X coord\nfor each ring", shape="box", style="rounded"
    )
    dot.node(name="is_more_lane", label="Is there another lane?", shape="diamond")
    dot.node(name="is_more_lane_yes", label="Yes.")
    dot.node(name="is_more_lane_no", label="No.")
    dot.node("end", shape="doublecircle", label="End")

    dot.edge("start", "is_tournament_created")
    dot.edges(
        [
            ("is_tournament_created", "is_tournament_created_yes"),
            ("is_tournament_created", "is_tournament_created_no"),
        ]
    )
    dot.edge("is_tournament_created_no", "create_tournament")
    dot.edge("create_tournament", "set_lane")
    dot.edge("set_lane", "get_target_hight", weight="1")
    dot.edge("set_lane", "get_x_rings", weight="1")
    dot.edge("set_lane", "get_lane_length", weight="1")
    dot.edge("set_lane", "is_more_lane", weight="5")
    dot.edges([("is_more_lane", "is_more_lane_yes"), ("is_more_lane", "is_more_lane_no")])
    dot.edge("is_more_lane_yes", "set_lane", weight="5")
    dot.edge("is_more_lane_no", "end", weight="5")
    dot.edge("is_tournament_created_yes", "end", weight="5")
    dot.render("tournament_creation", **opt)


def archer_registration(**opt: Mapping[str, Any]) -> None:
    dot = Digraph(
        comment="Archer registration flow.",
        node_attr={"color": "blue"},
        edge_attr={"color": "blue"},
        graph_attr={
            "label": "Archer registration flow.",
            "labelloc": "t",
            "labeljust": "c",
            "fontsize": "20",
        },
    )
    dot.node("start", shape="circle", label="Start")
    dot.node(name="is_archer_reg", label="Is the archer registered?", shape="diamond")
    dot.node(name="archer_exist", label="Does the archer exist?", shape="diamond")
    dot.node(name="is_archer_reg_yes", label="Yes.")
    dot.node(name="is_archer_reg_no", label="No.")
    dot.node(name="archer_exist_yes", label="Yes.")
    dot.node(name="archer_exist_no", label="No.")
    dot.node(name="ready", label="Ready to participate\nin a tournament.", shape="box")
    dot.node(name="register", label="Register the archer\nin a tournament.", shape="box")
    dot.node(name="collect_info", label="Collect info from the archer.", shape="box")
    dot.node("end", shape="doublecircle", label="End")
    dot.edge("start", "is_archer_reg")
    dot.edges([("is_archer_reg", "is_archer_reg_yes"), ("is_archer_reg", "is_archer_reg_no")])
    dot.edge("is_archer_reg_no", "archer_exist")
    dot.edges([("archer_exist", "archer_exist_yes"), ("archer_exist", "archer_exist_no")])
    dot.edge("archer_exist_no", "collect_info")
    dot.edge("collect_info", "register")
    dot.edge("archer_exist_yes", "register")
    dot.edge("is_archer_reg_yes", "ready")
    dot.edge("register", "ready")
    dot.edge("ready", "end")
    dot.render("archer_registration", **opt)


def main() -> None:
    docs_path = Path(__file__).parent
    opt: Mapping[str, Any] = {
        "overwrite_source": True,
        "format": "png",
        "view": True,
        "cleanup": True,
        "directory": docs_path,
    }
    archer_registration(**opt)
    tournament_creation(**opt)
    shooting_flow(**opt)


if __name__ == "__main__":
    main()
