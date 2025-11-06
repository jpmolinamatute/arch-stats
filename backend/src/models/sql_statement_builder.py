"""Utilities to build PostgreSQL statements with basic safety checks.

This module assembles SQL strings and performs light validation on
identifiers and conditions to reduce SQL-injection risk. Use these helpers
with asyncpg prepared statements so that values are always bound separately
from the SQL text.
"""

from string import Template


class SQLStatementBuilder:
    """Helper to construct simple, validated SQL for a single table.

    The builder:
    - Uses positional placeholders ($1, $2, ...) for values.
    - Validates identifiers and WHERE conditions to minimize injection risk.
    - Emits compact SQL compatible with asyncpg.

    Notes:
    - ``table_name`` must be a trusted identifier (not user input). It is
      interpolated directly in SQL.
    - Callers are responsible for consistent placeholder numbering across
      clauses and for binding values via prepared statements.
    """

    def __init__(self, table_name: str) -> None:
        """Initialize a builder for the given table name.

        Args:
            table_name: Trusted table identifier. May be schema-qualified,
                e.g. "public.archers". Do not pass user input here.

        """
        self.table_name = table_name
        self.select_template = Template(
            f"""
                SELECT $columns
                FROM {self.table_name}
                $where_clause
                $and_clauses
                ORDER BY created_at ASC
                $limit_clause;
            """
        )
        self.insert_template = Template(
            f"""
                INSERT INTO {self.table_name} ($columns)
                VALUES ($values)
                RETURNING {self.table_name}_id;
            """
        )
        self.update_template = Template(
            f"""
            UPDATE {self.table_name}
            SET $set_clause
            WHERE $where_clause
            $and_clauses;
        """
        )
        self.execute_select_function_template = Template(
            """
                SELECT * FROM $function_name($placeholders);
            """
        )
        self.select_view_template = Template(
            """
                SELECT $columns
                FROM $view_name
                $where_clause
                $and_clauses
                $order_by_clause
                $limit_clause;
            """
        )
        self.delete_template = Template(
            f"""
                DELETE FROM {self.table_name}
                WHERE $condition
                $and_clauses;
            """
        )

    def check_placeholder(self, value: str) -> bool:
        """Return True if ``value`` looks like a positional placeholder.

        A valid placeholder starts with ``$`` and is followed by one or more
        digits (e.g., ``$1``, ``$2``).

        Args:
            value: String to check.

        Returns:
            True if the string is a placeholder; False otherwise.
        """
        return value.startswith("$") and value[1:].isdigit()

    def validate_condition(self, condition: str) -> bool:
        """Validate a simple WHERE condition for safety.

        Conditions must have the form ``<column> <operator> <rhs>`` where
        ``<column>`` is a simple identifier (``str.isidentifier()``), and the
        operator is one of a small allowlist. The right-hand side must be a
        positional placeholder (e.g., ``$1``) except for ``IS``/``IS NOT``,
        which allow ``TRUE``, ``FALSE``, ``UNKNOWN``, or ``NULL``.

        This is a conservative, best-effort validation to reduce injection
        risks. It is not a full SQL parser.

        Args:
            condition: Raw condition string, e.g. ``"score >= $1"`` or
                ``"deleted_at IS NULL"``.

        Returns:
            True if the condition passes validation; False otherwise.
        """
        # a condition is a string like "column = $1" or "column >= $2". It MUST be column name
        # followed by operator and then followed by a placeholder or NULL.
        # basic validation to prevent SQL injection
        allowed_operators = {
            "IS NOT DISTINCT FROM",
            "IS DISTINCT FROM",
            "IS NOT",
            "IS",
            "<=",
            ">=",
            "<>",
            "LIKE",
            "IN",
            "=",
            "<",
            ">",
        }
        # Iterate longest operators first to avoid partial matches.
        # Example: avoid matching "IS" inside "IS NOT DISTINCT FROM".
        for operator in sorted(allowed_operators, key=len, reverse=True):
            if operator not in condition:
                continue

            parts = condition.split(operator)
            if len(parts) != 2:
                return False

            column, value = parts
            column = column.strip()
            value = value.strip()

            # Column must be a simple identifier (snake_case, etc.)
            if not column.isidentifier():
                return False

            if operator in {"IS", "IS NOT"}:
                # Allow explicit boolean/unknown/null checks for IS/IS NOT
                # e.g., col IS TRUE | FALSE | UNKNOWN | NULL
                return value.upper() in {"TRUE", "FALSE", "UNKNOWN", "NULL"}

            return self.check_placeholder(value)
        return False

    def build_select_with_conditions(
        self, columns: list[str] | None = None, conditions: list[str] | None = None, limit: int = 0
    ) -> str:
        """Build a SELECT statement with optional WHERE and LIMIT.

        All conditions are validated with :py:meth:`validate_condition` to
        reduce injection risk.

        Args:
            columns: List of column names to select. If empty, uses ``*``.
            conditions: List of raw condition strings. The first becomes
                ``WHERE <cond>``, the rest are joined with ``AND``.
            limit: If greater than 0, appends ``LIMIT <limit>``.

        Returns:
            A SQL string like:
            ``SELECT col1, col2 FROM table WHERE a = $1 AND b >= $2
            ORDER BY created_at ASC LIMIT 10;``

        Raises:
            ValueError: If any condition fails validation.

        Example:
            columns = ["id", "score"]
            conditions = ["archer_id = $1", "score >= $2"]
            sql = builder.build_select_with_conditions(columns, conditions, 50)
        """
        columns_str = ", ".join(columns) if columns else "*"
        where_clause = ""
        and_clauses = ""
        limit_clause = f"LIMIT {limit}" if limit > 0 else ""
        if conditions:
            # Validate all conditions before constructing the query to avoid SQL injection patterns
            for condition in conditions:
                if not self.validate_condition(condition):
                    raise ValueError(f"Invalid SQL condition: {condition}")
            where_clause = "WHERE " + conditions[0]
            and_clauses = ""
            if len(conditions) > 1:
                and_clauses = " AND " + " AND ".join(conditions[1:])

        return self.select_template.substitute(
            columns=columns_str,
            where_clause=where_clause,
            and_clauses=and_clauses,
            limit_clause=limit_clause,
        )

    def build_simple_select(self) -> str:
        """Build ``SELECT * FROM <table> ORDER BY created_at ASC;``.

        Returns:
            The SQL string for a simple select without WHERE or LIMIT.
        """
        return self.build_select_with_conditions(
            columns=[],
            conditions=[],
            limit=0,
        )

    def build_insert(self, columns: list[str]) -> str:
        """Build an INSERT statement with positional placeholders.

        Args:
            columns: Column names to insert values into.

        Returns:
            SQL string like ``INSERT INTO table (c1, c2) VALUES ($1, $2)
            RETURNING table_id;``

        Notes:
            Assumes the primary key follows the ``<table>_id`` convention for
            the ``RETURNING`` clause.
        """
        values = ", ".join(f"${i}" for i in range(1, len(columns) + 1))

        return self.insert_template.substitute(
            columns=", ".join(columns),
            values=values,
        )

    def build_update(self, set_data: list[tuple[str, str]], conditions: list[str]) -> str:
        """Build an UPDATE statement that always requires a WHERE clause.

        Args:
            set_data: Pairs of ``(column, placeholder)`` for the SET clause.
                Example: ``[("score", "$1"), ("updated_at", "$2")]``.
            conditions: Validated conditions joined by ``AND``.

        Returns:
            SQL string like ``UPDATE table SET score = $1 WHERE id = $2;``

        Raises:
            ValueError: If no conditions are provided, a condition is invalid,
                a column name is invalid, or a placeholder is malformed.
        """
        if not conditions:
            raise ValueError("Update statement requires at least one condition for safety.")

        for condition in conditions:
            if not self.validate_condition(condition):
                raise ValueError(f"Invalid SQL condition: {condition}")

        for col, val in set_data:
            if not col.isidentifier():
                raise ValueError(f"Invalid column name: {col}")
            if not self.check_placeholder(val):
                raise ValueError(f"Invalid value placeholder: {val}")

        where_clause = conditions[0]
        and_clauses = ""
        if len(conditions) > 1:
            and_clauses = " AND " + " AND ".join(conditions[1:])

        return self.update_template.substitute(
            set_clause=", ".join(f"{col} = {val}" for col, val in set_data),
            where_clause=where_clause,
            and_clauses=and_clauses,
        )

    def build_select_function(self, function_name: str, no_placeholders: int = 0) -> str:
        """Build ``SELECT * FROM <function>($1, ...);``.

        Args:
            function_name: Trusted function identifier.
            no_placeholders: Number of positional placeholders to include.

        Returns:
            SQL string that calls the function, e.g. ``SELECT * FROM f($1);``
            or ``SELECT * FROM f();`` when ``no_placeholders`` is 0.
        """
        placeholders = ", ".join(f"${i}" for i in range(1, no_placeholders + 1))
        return self.execute_select_function_template.substitute(
            function_name=function_name, placeholders=placeholders
        )

    def build_select_view(
        self,
        view_name: str,
        order_by_clause: str = "",
        columns: list[str] | None = None,
        conditions: list[str] | None = None,
        limit: int = 0,
    ) -> str:
        """Build a SELECT statement against a view with optional WHERE and LIMIT.

        All conditions are validated with :py:meth:`validate_condition` to
        reduce injection risk.

        Args:
            view_name: Trusted view identifier.
            columns: List of column names to select. If empty, uses ``*``.
            conditions: List of raw condition strings. The first becomes
                ``WHERE <cond>``, the rest are joined with ``AND``.
            limit: If greater than 0, appends ``LIMIT <limit>``.

        Returns:
            A SQL string like:
            ``SELECT col1, col2 FROM view WHERE a = $1 AND b >= $2
            ORDER BY created_at ASC LIMIT 10;``
        """
        columns_str = ", ".join(columns) if columns else "*"
        where_clause = ""
        and_clauses = ""
        limit_clause = f"LIMIT {limit}" if limit > 0 else ""
        if conditions:
            # Validate all conditions before constructing the query to avoid SQL injection patterns
            for condition in conditions:
                if not self.validate_condition(condition):
                    raise ValueError(f"Invalid SQL condition: {condition}")
            where_clause = "WHERE " + conditions[0]
            and_clauses = ""
            if len(conditions) > 1:
                and_clauses = " AND " + " AND ".join(conditions[1:])

        return self.select_view_template.substitute(
            view_name=view_name,
            columns=columns_str,
            where_clause=where_clause,
            and_clauses=and_clauses,
            order_by_clause=order_by_clause,
            limit_clause=limit_clause,
        )

    def build_delete(self, conditions: list[str]) -> str:
        if not conditions:
            raise ValueError("Update statement requires at least one condition for safety.")

        for condition in conditions:
            if not self.validate_condition(condition):
                raise ValueError(f"Invalid SQL condition: {condition}")
        where_clause = conditions[0]
        and_clauses = ""
        if len(conditions) > 1:
            and_clauses = " AND " + " AND ".join(conditions[1:])
        return self.delete_template.substitute(where_clause=where_clause, and_clauses=and_clauses)
