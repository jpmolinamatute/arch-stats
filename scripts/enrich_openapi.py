#!/usr/bin/env python
"""
Enrich OpenAPI 3.0 specification for frontend TypeScript code generation compatibility.

This script post-processes the OpenAPI 3.0 JSON specification generated from
swaggo (Swagger 2.0 -> OpenAPI 3.0 via swagger2openapi) to ensure complete
compatibility with frontend expectations:
1. Adds schema aliases (ArcherId, BowStyleType, GenderType, SlotLetterType).
2. Adds AuthNeedsRegistration, ValidationError, and HTTPValidationError schemas.
3. Sets nullable: true on pointer / optional domain model fields.
"""

import copy
import json
import sys
from pathlib import Path


def enrich_openapi(spec_path: Path) -> None:
    with spec_path.open("r", encoding="utf-8") as f:
        spec = json.load(f)

    schemas = spec.setdefault("components", {}).setdefault("schemas", {})

    # 1. Compatibility schema aliases
    alias_mappings = {
        "ArcherId": "ArcherID",
        "BowStyleType": "Bowstyle",
        "GenderType": "Gender",
        "SlotLetterType": "SlotLetter",
    }
    for alias, target in alias_mappings.items():
        if target in schemas and alias not in schemas:
            schemas[alias] = copy.deepcopy(schemas[target])

    # 2. AuthNeedsRegistration schema
    if "AuthNeedsRegistration" not in schemas:
        schemas["AuthNeedsRegistration"] = {
            "type": "object",
            "required": [
                "google_email",
                "google_subject",
                "given_name_provided",
                "family_name_provided",
            ],
            "properties": {
                "status": {
                    "$ref": "#/components/schemas/AuthStatus",
                    "default": "needs_registration",
                },
                "google_email": {
                    "type": "string",
                    "format": "email",
                    "title": "Google Email",
                },
                "google_subject": {
                    "type": "string",
                    "title": "Google Subject",
                },
                "given_name": {
                    "type": "string",
                    "nullable": True,
                    "title": "Given Name",
                },
                "family_name": {
                    "type": "string",
                    "nullable": True,
                    "title": "Family Name",
                },
                "given_name_provided": {
                    "type": "boolean",
                    "title": "Given Name Provided",
                },
                "family_name_provided": {
                    "type": "boolean",
                    "title": "Family Name Provided",
                },
                "picture_url": {
                    "type": "string",
                    "nullable": True,
                    "title": "Picture Url",
                },
            },
            "title": "Auth Needs Registration Response",
        }

    # 3. Validation error schemas
    if "ValidationError" not in schemas:
        schemas["ValidationError"] = {
            "type": "object",
            "required": ["loc", "msg", "type"],
            "properties": {
                "loc": {
                    "type": "array",
                    "items": {"anyOf": [{"type": "string"}, {"type": "integer"}]},
                },
                "msg": {"type": "string"},
                "type": {"type": "string"},
            },
        }

    if "HTTPValidationError" not in schemas:
        schemas["HTTPValidationError"] = {
            "type": "object",
            "properties": {
                "detail": {
                    "type": "array",
                    "items": {"$ref": "#/components/schemas/ValidationError"},
                },
            },
        }

    # 4. Nullable fields mapping for pointer / optional domain fields
    nullable_fields = {
        "ArcherCreate": ["club_id", "google_picture_url"],
        "ArcherFilter": [
            "archer_id",
            "first_name",
            "last_name",
            "gender",
            "bowstyle",
            "last_login_at",
            "created_at",
            "draw_weight",
            "club_id",
            "google_subject",
        ],
        "ArcherRead": ["club_id", "google_picture_url"],
        "ArcherSet": [
            "first_name",
            "last_name",
            "gender",
            "bowstyle",
            "google_picture_url",
            "last_login_at",
            "draw_weight",
            "club_id",
        ],
        "AuthRegistrationRequest": ["first_name", "last_name", "club_id"],
        "FullSlotInfo": ["club_id", "shot_per_round"],
        "SessionId": ["session_id"],
        "SessionRead": ["closed_at"],
        "ShotCreate": ["x", "y", "score", "arrow_id", "created_at"],
        "ShotRead": ["x", "y", "score", "arrow_id"],
        "SlotJoinRequest": ["club_id", "shot_per_round"],
    }

    for s_name, fields in nullable_fields.items():
        if s_name in schemas:
            props = schemas[s_name].get("properties", {})
            for fld in fields:
                if fld in props:
                    props[fld]["nullable"] = True

    with spec_path.open("w", encoding="utf-8") as f:
        json.dump(spec, f, indent=2)


def validate_file_path(raw: str) -> Path:
    """Validate that *raw* refers to an existing regular file and return its Path."""
    p = Path(raw)
    if not p.exists():
        print(f"Error: file not found: {p}", file=sys.stderr)
        sys.exit(1)
    if not p.is_file():
        print(f"Error: not a regular file: {p}", file=sys.stderr)
        sys.exit(1)
    return p


def main() -> None:
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <path_to_openapi.json>", file=sys.stderr)
        sys.exit(1)
    spec_path = validate_file_path(sys.argv[1])
    enrich_openapi(spec_path)


if __name__ == "__main__":
    main()
