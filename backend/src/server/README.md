# Arch-Stats Backend Server

*Who this is for:* Python developers working on the Arch-Stats FastAPI backend (API server and database layer).

This document describes the structure and conventions of the backend server code (in `backend/src/server/`), including how requests are handled via FastAPI routers, how data schemas are defined with Pydantic v2, how database operations use asyncpg, and how to run and extend tests.

## Request -> Validation -> DB -> Response: Overview

The Arch-Stats backend is an **async** web API built with **FastAPI**. It exposes REST endpoints under `/api/v0/` and uses **PostgreSQL** (via the asyncpg driver) for data storage. The flow for handling a request is:

1. **Routing:** FastAPI routes are defined in modular router files (e.g., `sessions_router.py`, `arrows_router.py` in the [`routers/`](./routers) package). Each router corresponds to a resource (Sessions, Arrows, Shots, Targets, etc.). For example, the [`sessions_router.py`](./routers/sessions_router.py) defines endpoints for creating, updating, and retrieving shooting sessions.
2. **Validation with Pydantic:** Request payloads and query parameters are defined using **Pydantic v2** models. For instance, a POST to create a session expects a `SessionsCreate` model (defined in [`schema/sessions_schema.py`](./schema/sessions_schema.py)), and returns data in a `SessionsRead` model. FastAPI automatically deserializes and validates the incoming JSON against these models. If validation fails, a 422 response is returned before hitting our logic.
3. **Database Operation:** Inside the route function, a database handler object (e.g., `SessionsDB`) is used to perform the needed data operation. These handlers are defined in the [`models/`](./models) package and use an async **connection pool** (via asyncpg) to query the PostgreSQL database. We do **not** use an ORM; instead, simple SQL statements are constructed and executed with asyncpg, and results are mapped back into Pydantic models.
4. **Response Construction:** The route returns either the Pydantic result or an error. We wrap responses in a unified structure using an `HTTPResponse` model (see [`routers/utils.py`](./routers/utils.py)). This `HTTPResponse` is a generic Pydantic model that contains a status code, either a `data` field (with the result data) or an `errors` list. A helper function `db_response()` runs the DB operation, catches any exceptions (like `DBNotFound` or `DBException` from our DB layer), and returns an appropriate JSONResponse. This ensures the API always returns a consistent JSON shape, e.g.:

   ```json
   { "code": 200, "data": { ... }, "errors": [] }
   // or an error:
   { "code": 404, "data": null, "errors": ["sessions: No record found with ..."] }
   ```

Example: Creating a new Session
When a client POST /api/v0/session with JSON body, FastAPI uses the SessionsCreate model to validate the input. Our router function might look like:

   ```python
   @SessionsRouter.post("", response_model=HTTPResponse[None])
   async def create_session(
       session_data: SessionsCreate,
       sessions_db: SessionsDB = Depends(get_sessions_db),
   ) -> JSONResponse:
       return await db_response(sessions_db.insert_one, status.HTTP_201_CREATED, session_data)
   ```

Here's what happens:

session_data: SessionsCreate - FastAPI gives us a SessionsCreate instance (already validated).

sessions_db = Depends(get_sessions_db) - FastAPI provides a SessionsDB instance (the dependency opens a db connection from the pool).

We call db_response(sessions_db.insert_one, 201, session_data). The insert_one method (defined in our DBBase class) will insert a new row into the sessions table using asyncpg and return a SessionsRead Pydantic model of the new session.

db_response catches any database exceptions:

If insertion fails (e.g., constraint violation), a DBException is raised and db_response returns an HTTP 400 with error.

On success, insert_one returns a SessionsRead. db_response wraps it in HTTPResponse with code 201 and the FastAPI endpoint sends that JSON to the client.

FastAPI also automatically generates an OpenAPI spec (available at /api/openapi.json and interactive docs at /api/swagger) from these models and route definitions.

## Pydantic v2 Models (Schemas)

All request and response schemas are defined with Pydantic v2 (which uses BaseModel). They live in the schema/ package, organized by resource. For example:

sessions_schema.py defines:

SessionsCreate - required fields to create a session (e.g., location: str, start_time: datetime, etc.). Extra fields are forbidden (model_config = ConfigDict(extra="forbid") to catch unexpected input).

SessionsUpdate - optional fields for patching a session (all fields are Optional[...]). This uses extra="forbid" and populate_by_name=True (so we can accept either field names or aliases if defined).

SessionsRead - fields for reading a session back from the API. It extends SessionsCreate and adds the session_id: UUID (with alias "id") and possibly computed fields like end_time. It also can include helper methods (e.g., get_id() method to retrieve the UUID).

SessionsFilters - for query params, often just an alias of SessionsUpdate (to reuse the optional fields for filtering).

Similarly, shots_schema.py defines ShotsCreate, ShotsRead, etc. and so on for arrows and targets.

Strict typing & validation: We enable strict validation where appropriate (for example, extra="forbid" means if a client sends any unknown field, it will be rejected). Type hints in Pydantic models are used extensively (e.g., distance: int ensures we only accept integers for the distance field). Pydantic v2's new features like model_dump and model_copy are used in the code for serialization and cloning models (for instance in tests or when converting to JSON for the database).

All Pydantic models are type-hinted and we run mypy in CI to enforce correctness. The use of Pydantic ensures that by the time data reaches our database layer, it's already validated and correctly typed (e.g., dates are Python datetime objects, UUIDs are UUID instances, etc.).

## Database Layer: asyncpg and DBBase

Database operations are implemented in the models/ module. Key points:

We maintain a connection pool using asyncpg. DBPool (in server/db_pool.py) encapsulates asyncpg.create_pool and provides get_db_pool() to retrieve the singleton pool. The pool is initialized when the FastAPI app starts (see the lifespan function in app.py), and closed on shutdown.

There's a generic base class DBBase which implements common CRUD logic using type parameters. It's defined as DBBase[CreateModel, UpdateModel, ReadModel] and provides methods like insert_one(data: CreateModel) -> UUID, get_one_by_id(id) -> ReadModel, get_all(filter_dict) -> list[ReadModel], update_one(id, data: UpdateModel) -> None, and delete_one(id) -> None. Internally, it constructs SQL strings and uses asyncpg's prepared statements (with $1, $2 placeholders) to execute queries.

DBBase._convert_row converts an asyncpg Record to a Pydantic ReadModel (our ReadModel classes implement a .get_id() and Pydantic handles field population including alias like id).

It also handles JSON fields by serializing Python dicts/lists to JSON strings when inserting (see _serialize_db_value).

If a query finds no results for a given ID or filter, DBBase raises a custom DBNotFound exception; other database errors raise DBException. These exceptions propagate up to the router where db_response turns them into HTTP 404 or 400 responses respectively.

Specific model classes (e.g., SessionsDB, ShotsDB, etc.) extend DBBase. In their constructor, they define the SQL schema for their table and call super().**init** with the table name and schema. For example, ShotsDB.**init** provides the table name "shots" and a multiline string for the table schema (columns, types, constraints) then calls the base initializer. This allows us to programmatically create or drop tables (useful in testing and initial startup).

Model classes can also define resource-specific methods. For instance, ShotsDB adds a method create_notification(channel: str) that sets up a Postgres trigger to notify on new shots (used for real-time updates). It also overrides some base behaviors when needed (like a helper get_by_session_id to get all shots for a session).

asyncpg usage: All DB operations use async/await with the asyncpg Pool. We use pool.acquire() context managers to get a connection for a short operation, and we never hold locks for long. This means multiple requests can run concurrently on different connections from the pool. SQL statements are kept simple (CRUD with basic filtering). More complex logic (like ensuring only one open session at a time) is enforced either at the application level or via database constraints/triggers.

Because we don't use an ORM, the code is straightforward to follow: each method in DBBase builds an SQL string and executes it. This maximizes performance (asyncpg is very fast for bulk operations and uses prepared statements internally) and keeps typing strict (no magic attributes - everything is explicit).

## FastAPI App Configuration

The FastAPI application is created in app.py. Key aspects:

We use FastAPI(debug=settings.arch_stats_dev_mode, lifespan=lifespan, title="Arch Stats API", version="0.1.0", openapi_url="/api/openapi.json", docs_url="/api/swagger"). The lifespan function is defined to handle startup/shutdown events asynchronously.

Startup (lifespan): On startup, we call DBPool.create_db_pool() to initialize the asyncpg pool (using env vars from .env for DB credentials). We then automatically call manage_tables("create") if running in debug mode - this creates all database tables (if they don't exist) and in the case of Shots, sets up the notification trigger. This is mainly for development convenience. (In production, debug=False, we might skip auto-creation or use migrations instead.)

Shutdown: The lifespan context ensures that on shutdown, if we were in debug mode and a CancelledError triggers (e.g., Ctrl+C during dev), it will drop tables for a clean slate. Finally it closes the DB pool.

We mount the frontend at runtime: app.mount("/app", StaticFiles(directory=.../frontend, html=True), name="frontend"). This serves the built frontend files under the URL prefix /app. In production after running npm run build, the static index.html and assets live in backend/src/server/frontend/. So a user accessing the root of the server (e.g. /app/) will get the UI.

Routers: We include all API routers with the prefix /api/v0. The prefix includes a version (v0) which is derived from the app version. The included routers are SessionsRouter, ArrowsRouter, ShotsRouter, TargetsRouter, and WSRouter (websocket routes). Each corresponds to a module in routers/. The WebSocket router (WSRouter in websocket.py) sets up endpoints for clients to subscribe to real-time events (specifically, it likely listens for Postgres NOTIFY messages on the channel defined by ARCH_STATS_WS_CHANNEL and pushes new shot data to frontend).

## Strict Typing and Conventions

This project is configured for strict type checking:

We use Python type hints pervasively. For example, database methods are generically typed to return the correct Pydantic model type. The custom Protocol HasId ensures our Read models have a get_id() method for the generic constraint.

Mypy is run in CI with a configuration in pyproject.toml to ensure no type regressions.

Pylint and Black/Isort enforce code style and import order. The project follows PEP8/PEP257 standards by using these tools (check pyproject.toml for their settings). All CI linting must pass before merging.

We prefer explicitness: for example, all SQL statements are written out (with proper parameterization to avoid SQL injection). There is a conscious trade-off of writing a bit more code for clarity and type-safety.

If you introduce new dependencies or modules, add type hints and update the pyproject config for mypy if needed. The goal is that running pytest and pylint/mypy locally should produce zero errors or warnings.

## Testing (pytest + asyncio)

Tests for the backend reside in the backend/tests/ directory, separated by domain (tests/endpoints/ and tests/models/). We use pytest with pytest-asyncio for async test support:

In tests/conftest.py, we define fixtures for setting up the FastAPI app and database for tests. Notably, the async_client fixture uses FastAPI's ASGI lifespans to startup and teardown the app within the test, and uses httpx.AsyncClient with ASGITransport to simulate HTTP calls in-memory (no actual network calls). This means we run the server internally and can call endpoints as if the server were live.

We also have a db_pool_initialed fixture that creates the DB pool and ensures tables are created before a test (and drops them after). This is used in model tests where tests directly call methods like SessionsDB.insert_one without going through the API.

Database isolation: Tests currently run against the same database schema, but each test that uses db_pool_initialed will drop tables at the start (and end) to avoid leftover data. When running pytest, the manage_tables("create") and "drop" calls ensure a clean state. The database is the same Postgres instance used for dev, so be cautious - running tests will wipe any data in those tables. It's recommended to use a dedicated test database name or a disposable Docker container. (By default, if you use the provided Docker setup via manage_docker, it's pointing to a transient container.)

Running tests: Activate the virtual environment (source backend/.venv/bin/activate after an uv sync) and simply run pytest in the backend/ directory. The test suite will start a Postgres via Docker (if not already up) and execute all tests. All tests are marked @pytest.mark.asyncio since they are async. The CI does not currently run tests automatically (only linters), so it's important to run tests locally or incorporate them into CI if needed.

Example test flow: In tests/endpoints/test_sessions_endpoints.py, there are tests like test_create_session which uses the async_client fixture to POST to the /api/v0/session endpoint with some JSON. The response is then asserted to have code == 201. The test fixture ensures that before the request, the database tables are created and after the test they are dropped, so each test is isolated.

## File Structure Reference

For easy navigation, here are important backend files and their roles:

Entry & Config:

app.py - Creates the FastAPI app, sets up routers and static files, and manages startup/shutdown (DB pool and table setup).

settings.py - Uses pydantic settings management to load environment variables (like DB credentials, server port, debug mode). This feeds into the app and DB config.

## Database Layer

db_pool.py - Singleton pattern for asyncpg pool (connects using settings from .env or defaults).

models/base_db.py - Generic database access layer (CRUD operations and utilities).

models/ (e.g., sessions_db.py, arrows_db.py, etc.) - Concrete DB classes for each table. They also define the SQL schema (used by create_table on startup or in tests).

models/queries.sql - (If present) Might contain raw SQL queries or migration scripts (not heavily used if we rely on the schema in code).

Schema (Data Models):

schema/ - Pydantic models for input/output. E.g., SessionsCreate, SessionsRead, etc. and any additional complex types (Targets schema might include a submodel for target face configuration).

## Routers (API Endpoints)

routers/ - Each file is an APIRouter instance:

sessions_router.py, arrows_router.py, shots_router.py, targets_router.py - define REST endpoints for those resources.

websocket.py - sets up a WebSocket route (likely at /api/v0/ws or similar) to stream real-time updates (the server may listen to Postgres NOTIFY and push to clients).

utils.py - defines the HTTPResponse model and db_response helper for consistent API responses.

Shared & Utilities:

shared/ - This may contain code shared across server and other parts (e.g., factories for test data in shared/factories/ used by tests to generate dummy arrows, sessions, etc.).

[../arrow_reader/, ../target_reader/, etc.] - (If present adjacent to server/) might contain separate services (for hardware integration, etc.). The server communicates with those via API or internal calls, but those are outside the scope of the server's direct request/response cycle.

## Extending and Maintaining

- Adding a new endpoint:Create a new route in an existing router or a new router module, and include it in app.py.
- Validate data in and out: Define a Pydantic schema for request/response in the schema module.  Use the existing patterns: dependency-inject a DB class instance, call its method, and wrap with db_response.
- Adding a new model/table: Create a new XyzDB class in models, extending DBBase with appropriate type parameters and table schema. Register any custom methods or triggers needed. Add an instance of it in app.lifespan.manage_tables("create") if you want auto-creation in dev. Write a Pydantic schema for it (XyzCreate/Read/Update) and expose via a router.

By following the established patterns (strict Pydantic models, clearly defined DB methods, and consistent response wrapping), you can implement new features confidently. Always run tests (pytest) after changes and ensure type checks pass (mypy). The combination of type safety and tests will catch most issues early in development. Happy coding!
