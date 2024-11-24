# The Server

We will be using FastAPI, SQLAlchemy, Pydantic, and Uvicorn, all compiled using Cython. First, we need to create SQLAlchemy schemas in Python (.pyx files) for registration, tournament, lane, archer, target, arrow, and shots. Next, we will create an endpoint to serve the WebUI. For each of the entities (registration, tournament, lane, archer, target, and arrow), we will implement GET, POST, PATCH, and DELETE endpoints. The shots will be created by the shot_reader service. Finally, we need to determine if we should create DELETE and PATCH endpoints for the shots.
