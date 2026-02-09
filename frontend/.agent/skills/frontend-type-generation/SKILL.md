---
name: frontend-type-generation
trigger: glob
description: CRITICAL. Use this rule whenever backend models change or when "generating types" for the frontend.
globs: frontend/**/*
---

# Frontend Type Generation & Sync

Keeping data types synced between the backend and frontend is a high priority.
Follow this logic exactly:

1. **Change Directory:** Always start by moving into the `./frontend/` directory.
2. **Connectivity Check:** Check if the backend server is currently active by running:
   `pgrep -af uvicorn`
3. **Execution Logic:**
   * **If Backend is Running:** Execute `npm run generate:types`.
   * **If Backend is NOT Running:** Execute the fallback script: `./scripts/generate_fe_types.bash`.
4. **Verification:** Confirm that the generated type files have been updated in the
   frontend source tree.
