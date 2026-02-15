---
name: frontend-type-generation
description: CRITICAL. How to keep frontend data types in sync with the backend, ensuring type safety and reducing bugs.
---

# Frontend Type Generation & Sync

Keeping data types synced between the backend and frontend is a high priority.
Follow this logic exactly:

1. **Change Directory:** Always start by moving into the `./frontend/` directory.
2. **Execute Generation:** Run the npm script:
   `npm run generate:types`
   *Note*: This script automatically handles checking if the backend is running and choosing the
   appropriate generation method.
3. **Verification:** Confirm that the generated type files have been updated in
   `./frontend/src/types/types.generated.ts`.
4. **Application:** Ensure that the new types are being used in the frontend codebase to
   maintain type safety and reduce bugs.
