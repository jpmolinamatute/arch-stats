# Live data ideas

## Stat data

```python
from uuid import UUID
from pydantic import BaseModel, Field


class Shot(BaseModel):
    shot_id: UUID = Field(...)
    score: int = Field(...)

class Stat(BaseModel):
    shots: list[Shot]  = Field(...) # this list MUST be sorted by created_at
    number_of_shots: int = Field(...)
    total_score: int = Field(...)
    max_score: int = Field(...)
    mean: float  = Field(...)
```

## Ways to get data

### WebSocket

It will push Stat data (this will include only the last shot). This is how it's going to accomplish it:

1. Listen to shot_insert_{slot_id}.
2. query `SELECT * from live_stat_by_slot_id WHERE slot_id = {p_slot_id}
3. Instantiate a Stat class using the result of step 1 as shots and `number_of_shots`, `total_score`, `max_score`, and `mean` from step 2.
4. Push the Stat object to the client.

### GET endpoint

it will return Stat data (this will include all shots)

```sql
SELECT *
FROM shot
WHERE slot_id = {p_slot_id}
ORDER BY created_at ASC;
```

## Frontend

dynamic table: each column represent a scoring shot and each row represent a round (archers will define the number of shot per round. they will also allow to change it any time) and the last column will be the sum all scoring shot for that specific row (aka round)

we are going to display the stats below the table and underneath of the stats we are going to display charts graph displaying all scoring so that the archer can see their performance during the shooting session.

## Plan

### Small Refactor

we currently rely on archer_id AND session_id, however this information is already stored in the slot table. At the same time during the live session, we will also need information on the slot table any ways. This is why we need to lean more on this table and use slot_id as much as possible.

1. Add shot_per_round column to the slot table. DONE
2. Update PostgreSQL function to use slot_id instead of session_id & archer_id DONE
3. Create a PostgreSQL function that accept an archer_id as param joins/query tables session.is_opened = true AND slot.is_shooting = true and returns the slot_id DONE
4. Update Session schema (add shot_per_round) DONE
5. Update slot model to call PostgreSQL function using slot_id instead of session_id & archer_id DONE
6. Update Slot tests / create new ones DONE
7. update WebUI (aka frontend)

### Shot implementation

#### Backend implementation

1. Create POST /api/v0/shot endpoint # add a new row in the shot table DONE
2. Create GET /api/v0/shot/by-slot/{slot_id} # pull all shot by slot_id AND sorted by created_at DONE
3. Come up with tests for these endpoints DONE
4. Finish the Stat class  DONE
5. Create WebSocket /ws/v0/shot STARTED

#### Frontend implementation

1. We should save slot information
2. Incorporate the `TargetFace.vue` in /app/live-session (after the session and slot are created). This means that if the page is reloaded or the browser tab is closed and then visited again, TargetFace.vue must be rendered. This replace the current behavior of /app/live-session.
3. Extending `TargetFace.vue` we are going to

## Questions

Where do we store the number of shot per round (aka how big the round is). This is only a reference to display the dynamic table for a specific shooting session
