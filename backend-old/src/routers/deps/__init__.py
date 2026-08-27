"""Shared dependencies for routers (auth and model providers).

Keep these lightweight and free of business logic. They should only wire
FastAPI dependencies (logger, db pool, auth) to concrete model classes.
"""
