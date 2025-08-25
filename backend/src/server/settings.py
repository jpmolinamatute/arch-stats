from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=str(object=Path(__file__).parent / ".env"))
    postgres_user: str = ""
    postgres_password: str = ""
    postgres_db: str = ""
    postgres_host: str = ""
    postgres_port: int = 5432
    postgres_socket_dir: str = ""
    arch_stats_dev_mode: bool = False
    arch_stats_server_port: int = 8000
    arch_stats_ws_channel: str = "archy"


settings = Settings()
