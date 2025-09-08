from uuid import UUID

from asyncpg import Pool
from pydantic import BaseModel

from shared.models.parent_model import ParentModel
from shared.schema import ShotsCreate, ShotsFilters, ShotsRead, ShotsUpdate
from shared.settings import settings


class ShotsModel(ParentModel[ShotsCreate, ShotsUpdate, ShotsRead, ShotsFilters]):
    def __init__(self, db_pool: Pool) -> None:
        super().__init__("shots", db_pool, ShotsRead)
        channel = settings.arch_stats_ws_channel
        self.func_name = f"notify_new_shot_{channel}"

    async def create(self) -> None:
        """Create the shots table, indexes, and NOTIFY trigger.

        - Columns: UUID PK, FKs to arrows and sessions, timing fields, optional x/y.
        - CHECK: (arrow_landing_time, x, y) are all present or all NULL.
        - Trigger: AFTER INSERT, NOTIFY the configured channel with NEW row JSON.
        """
        schema = """
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            arrow_id UUID NOT NULL,
            session_id UUID NOT NULL,
            arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrow_landing_time TIMESTAMP WITH TIME ZONE,
            x REAL,
            y REAL,
            FOREIGN KEY (arrow_id) REFERENCES arrows (id) ON DELETE CASCADE,
            FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
            CHECK (
                (
                    arrow_landing_time IS NOT NULL AND
                    x IS NOT NULL AND
                    y IS NOT NULL
                )
                OR
                (
                    arrow_landing_time IS NULL AND
                    x IS NULL AND
                    y IS NULL
                )
            )
        """.strip()
        channel = settings.arch_stats_ws_channel
        trigger_name = f"{self.func_name}_trigger"
        function_sql = f"""
        CREATE OR REPLACE FUNCTION {self.func_name}() RETURNS TRIGGER AS $$
        BEGIN
            PERFORM pg_notify('{channel}', row_to_json(NEW)::text);
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
        """.strip()
        trigger_sql = f"""
        DO $$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_trigger WHERE tgname = '{trigger_name}'
            ) THEN
                CREATE TRIGGER {trigger_name}
                AFTER INSERT ON shots
                FOR EACH ROW EXECUTE FUNCTION {self.func_name}();
            END IF;
        END$$;
        """.strip()
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")
            await conn.execute(
                f"CREATE INDEX IF NOT EXISTS idx_{self.name}_arrow_id ON {self.name} (arrow_id);"
            )
            await conn.execute(
                f"""CREATE INDEX IF NOT EXISTS idx_{self.name}_session_id
                ON {self.name} (session_id);"""
            )
            await conn.execute(function_sql)
            await conn.execute(trigger_sql)

    async def drop(self) -> None:
        """Drop indexes, table, trigger, and notify function created in create()."""
        trigger_name = f"{self.func_name}_trigger"
        async with self.db_pool.acquire() as conn:
            await conn.execute(f"DROP INDEX IF EXISTS idx_{self.name}_session_id;")
            await conn.execute(f"DROP INDEX IF EXISTS idx_{self.name}_arrow_id;")
            await conn.execute(f"DROP TABLE IF EXISTS {self.name};")
            await conn.execute(f"DROP TRIGGER IF EXISTS {trigger_name} on {self.name} ;")
            await conn.execute(f"DROP FUNCTION IF EXISTS {self.func_name};")

    async def update_one(self, _id: UUID, data: BaseModel) -> None:
        """Disallow updates because shots are write-protected in this model.

        Raises:
            NotImplementedError: Always.
        """
        raise NotImplementedError("Shots are write-protected; update not supported here.")

    async def insert_one(self, data: BaseModel) -> UUID:
        """Disallow direct creation because shots are write-protected here.

        Raises:
            NotImplementedError: Always.
        """
        raise NotImplementedError("Shots are write-protected; creation not supported here.")

    async def get_one_by_id(self, _id: UUID) -> ShotsRead:
        """Disallow single-record fetch by id in this model.

        Raises:
            NotImplementedError: Always.
        """
        raise NotImplementedError("Fetching a single shot is not supported.")
