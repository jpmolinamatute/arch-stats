from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=str(object=Path(__file__).parent / ".env"))
    postgres_user: str = Field(...)
    postgres_db: str = Field(...)

    postgres_port: int = Field(default=5432)
    postgres_password: str = Field(default="")
    postgres_host: str = Field(default="")
    postgres_socket_dir: str = Field(default="")
    arch_stats_dev_mode: bool = Field(default=False)
    arch_stats_server_port: int = Field(default=8000)
    arch_stats_ws_channel: str = Field(default="archy")


settings = Settings()
