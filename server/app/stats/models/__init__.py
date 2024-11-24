from contextlib import contextmanager
from typing import Generator

from sqlalchemy import create_engine
from sqlalchemy.orm import Session, declarative_base, sessionmaker

from app.stats.config import conf


DbSession = Generator[Session, None, None]
Base = declarative_base()


@contextmanager
def get_session() -> DbSession:
    user = conf.postgres_user
    password = conf.postgres_password
    name = conf.postgres_db
    host = conf.db_host
    port = conf.db_port
    database_url = f"postgresql+psycopg://{user}:{password}@{host}:{port}/{name}"
    engine = create_engine(database_url)
    session_local = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    session = session_local()
    try:
        yield session
    finally:
        session.close()


__all__ = ["Base"]
