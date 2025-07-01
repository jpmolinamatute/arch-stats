from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=str(object=Path(__file__).parent / ".env"))
    postgres_user: str = "juanpa"
    postgres_password: str = "changeme"
    postgres_db: str = "arch-stats"
    postgres_host: str = "localhost"
    postgres_port: int = 5432
    postgres_socket_dir: str = "/var/run/postgresql"
    arch_stats_dev_mode: bool = False
    arch_stats_server_port: int = 8000
    arch_stats_ws_channel: str = "archy"


settings = Settings()
