---
name: frontend-run-server
description: Use this to start the Vite development server for the Vue 3 frontend.
---

# Skill: Run Frontend Server

This skill automates launching the frontend development environment.

## Execution Steps

1. **Change Directory:** Navigate to the `./frontend/` directory.
2. **Execute Task:** Run the VS Code task: `"Start Vite Frontend"`.
   - *Note:* This task executes `npm run dev` within the `./frontend` context.
3. **Fallback:** If the task runner is unavailable, run manually:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Verification

- Confirm the Vite server is active (typically on <http://localhost:5173>).
- Ensure the frontend can communicate with the backend FastAPI service at /api/v0.
