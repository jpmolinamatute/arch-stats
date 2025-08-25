from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=str(object=Path(__file__).parent / ".env"))
    postgres_user: str = Field(..., description="Username used to connect to PostgreSQL")
    postgres_db: str = Field(..., description="PostgreSQL database name")

    postgres_port: int = Field(default=5432, description="PostgreSQL port")
    postgres_password: str = Field(
        default="",
        description="password used to connect to PostgreSQL in case not having POSTGRES_SOCKET_DIR",
    )
    postgres_host: str = Field(default="", description="PostgreSQL remote hostname or IP")
    postgres_socket_dir: str = Field(
        default="", description="PostgreSQL socket directory (preferred)"
    )
    arch_stats_dev_mode: bool = Field(
        default=False, description="Enable development mode for Archy Stats"
    )
    arch_stats_server_port: int = Field(default=8000, description="Port for the Archy Stats server")
    arch_stats_ws_channel: str = Field(
        default="archy", description="WebSocket channel for Archy Stats"
    )


settings = Settings()
