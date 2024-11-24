import glob
from setuptools import setup
from Cython.Build import cythonize
from pydantic import Field
from pydantic_settings import BaseSettings


class Configuration(BaseSettings):
    app_name: str = Field(...)
    postgres_user: str = Field(...)
    postgres_password: str = Field(...)
    postgres_db: str = Field(...)
    db_host: str = Field(...)
    db_port: str = Field(...)

    class Config:  # pylint: disable=R0903
        env_file = ".env"
        env_file_encoding = "utf-8"
        extra = "ignore"


py_files = glob.glob("**/*.py", recursive=True)
setup(ext_modules=cythonize(py_files, nthreads=4))

conf = Configuration()
