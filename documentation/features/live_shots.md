# Live data ideas

## Ways to get data

```python
from uuid import UUID
from pydantic import BaseModel, Field


class Shot(BaseModel):
    shot_id: UUID = Field(...)
    score: int = Field(...)

class Stat(BaseModel):
    shot: list[Shot]  = Field(...) # this list MUST be sorted by created_at
    count: int = Field(...)
    total_score: int = Field(...)
    max_score: int = Field(...)
    median: float = Field(...)
    mean: float  = Field(...)
```

### WebSocket

It will push Stat data (this will include only the last shot)

### GET endpoint

it will return Stat data (this will include all shots)

## Frontend

dynamic table: each column represent a scoring shot and each row represent a round (archers will define the number of shot per round. they will also allow to change it any time) and the last column will be the sum all scoring shot for that specific row (aka round)

we are going to display the stats below the table and underneath of the stats we are going to display charts graph displaying all scoring so that the archer can see their performance during the shooting session.

## Plan

### Small Refactor

we currently rely on archer_id AND session_id, however this information is already stored in the slot table. At the same time during the live session, we will also need information on the slot table. This is why we need to lean more on this table and only use slot_id.

1. Add shot_per_round column to the slot table.
2. Update PostgreSQL function to use slot_id instead of session_id & archer_id
3. Create a PostgreSQL function that accept an archer_id as param joins/query tables session.is_opened = true AND slot.is_shooting = true and returns the slot_id
4. Update Slot schema (add shot_per_round)
5. Update slot model to call PostgreSQL function using slot_id instead of session_id & archer_id
6. Update Slot tests / create new ones
7. update WebUI (aka frontend)

### Shot implementation

1. Create POST /api/v0/shot endpoint # add a new row in the shot table
2. Create GET /api/v0/shot/slot/{slot_id} # pull all shot by slot_id AND sorted by created_at
3. Come up with tests for these endpoints
4. Finish the Stat class
5. Create WebSocket /ws/v0/shot
6. Incorporate the `TargetFace.vue`

## Questions

Where do we store the number of shot per round (aka how big the round is). This is only a reference to display the dynamic table for a specific shooting session
