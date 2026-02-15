---
name: frontend-type-generation
description: CRITICAL. How to keep frontend data types in sync with the backend, ensuring type safety and reducing bugs.
---

# Frontend Type Generation & Sync

Keeping data types synced between the backend and frontend is a high priority.
Follow this logic exactly:

1. **Change Directory:** Always start by moving into the `./frontend/` directory.
2. **Connectivity Check:** Check if the backend server is currently active by running:
   `pgrep -af uvicorn`
3. **Execution Logic:**
    - **If Backend is Running:** Execute `npm run generate:types`.
    - **If Backend is NOT Running:** Execute the fallback script: `../scripts/generate_fe_types.bash`.
4. **Verification:** Confirm that the generated type files have been updated in ./frontend/src/types/types.generated.ts.
5. **Application:** Ensure that the new types are being used in the frontend codebase to
   maintain type safety and reduce bugs.
