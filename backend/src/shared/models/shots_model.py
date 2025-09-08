from uuid import UUID

from asyncpg import Pool
from pydantic import BaseModel

from shared.models.parent_model import ParentModel
from shared.schema import ShotsCreate, ShotsFilters, ShotsRead, ShotsUpdate
from shared.settings import settings


class ShotsModel(ParentModel[ShotsCreate, ShotsUpdate, ShotsRead, ShotsFilters]):

    def __init__(self, db_pool: Pool) -> None:
        super().__init__("shots", db_pool, ShotsRead)

    async def create(self) -> None:
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
        """
        channel = settings.arch_stats_ws_channel
        await self.execute(f"CREATE TABLE IF NOT EXISTS {self.name} ({schema});")
        await self.execute(
            f"CREATE INDEX IF NOT EXISTS idx_{self.name}_arrow_id ON {self.name} (arrow_id);"
        )
        await self.execute(
            f"CREATE INDEX IF NOT EXISTS idx_{self.name}_session_id ON {self.name} (session_id);"
        )
        await self.create_notification(channel)

    async def drop(self) -> None:
        await self.execute(f"DROP TABLE IF EXISTS {self.name} CASCADE;")

    async def create_notification(self, channel: str) -> None:
        """Ensure the notification function and trigger exist for new shots inserts."""

        function_sql = f"""
        CREATE OR REPLACE FUNCTION notify_new_shot_{channel}() RETURNS TRIGGER AS $$
        BEGIN
            PERFORM pg_notify('{channel}', row_to_json(NEW)::text);
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
        """
        trigger_sql = f"""
        DO $$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_trigger WHERE tgname = 'shots_notify_trigger_{channel}'
            ) THEN
                EXECUTE format(
                    'CREATE TRIGGER shots_notify_trigger_%s AFTER INSERT ON shots
                     FOR EACH ROW EXECUTE FUNCTION notify_new_shot_%s();',
                    '{channel}', '{channel}');
            END IF;
        END$$;
        """
        async with self.db_pool.acquire() as conn:
            await conn.execute(function_sql)
            await conn.execute(trigger_sql)

    async def update_one(self, _id: UUID, data: BaseModel) -> None:
        raise NotImplementedError("Shots are write-protected; update not supported here.")

    async def insert_one(self, data: BaseModel) -> UUID:
        raise NotImplementedError("Shots are write-protected; creation not supported here.")

    async def get_one_by_id(self, _id: UUID) -> ShotsRead:
        raise NotImplementedError("Fetching a single shot is not supported.")
