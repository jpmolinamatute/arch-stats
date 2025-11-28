# Arch-Stats Frontend

This is the Web App for the Arch-Stats project. We are using Vue3 with TypeScript.

## Prerequisites

- Node.js (v25.2.1 or higher)
- npm (v11.6.4 or higher)

## Mandatory steps

Before we run or build the Web App, we need to generate `./frontend/src/types/types.generated.ts` from the backend API. This file is essential to keep the backend and frontend in sync.

We have two ways to do this:

```bash
# On one terminal
./scripts/start_uvicorn.bash

# On another terminal
npm run generate:types

# After the types file is generated, you need to kill the uvicorn process, by pressing Ctrl+C on the first terminal.
```

The other way is to run

```bash
uv run "./scripts/generate_openapi.py"
cd "./frontend"
npx openapi-typescript "../openapi.json" --export-type --immutable --output "./src/types/types.generated.ts"
```

**Note**: The `/frontend/src/types/types.generated.ts` file is gitignored, and must NOT be committed to the repository. We also need to generate this file often, since there might be changes in the backend API that need to be reflected in the frontend.

## How to run the app

### In development

We use Vite to serve the web app in development. Run `npm run dev` to start the development server. The app will be available at <http://localhost:5173>. We also vite to proxy the API requests to the backend server running at <http://localhost:3000>.

### In production

We use Uvicorn to serve the web app and the API server. Run `uvicorn main:app --host 0.0.0.0 --port 8000` to start the production server. The whole app will be available at <http://localhost:8000>.

## Code quality

We use ESLint and Prettier to enforce code quality. Run `npm run lint` to lint the code, and `npm run format` to format the code. We also use Vitest to run tests. Run `npm run test` to run the tests. We also have `./scripts/linting.bash --lint-frontend` script to do all this and more. This script will:

- Generate types` file from the backend API.
- Lint the code.
- Format the code.
- Run tests.
- Build the app to make sure it works.

## Project Structure

The project is organized as follows:

- `src/api`: Contains the API client and configuration. This is where all backend communication happens.
- `src/composables`: Contains reusable business logic and state management using Vue's Composition API. We prefer composables over a global store for modularity.
- `src/components`: Reusable UI components. These should be dumb components as much as possible.
- `src/views`: Page-level components that connect composables with UI components.
- `src/types`: TypeScript definitions. `types.generated.ts` is auto-generated from the backend OpenAPI spec.

## Key Technologies & Decisions

- **Vue 3 & Composition API**: We use `<script setup>` syntax for all components.
- **TailwindCSS**: We use a custom configuration with World Archery (WA) brand colors.
  - Use `wa-reflex`, `wa-pink`, `wa-yellow`, etc., for brand colors.
  - Use semantic aliases like `primary`, `accent`, `success` for intent-driven styling.
- **Type Safety**: We rely heavily on `types.generated.ts` to ensure our frontend code matches the backend API. **Never manually edit this file.**

## Development Workflow

1. **New Features**:
    - Create a new branch.
    - If the backend API changed, run `npm run generate:types` (see Mandatory steps).
    - Implement logic in a Composable (if stateful) or Utility.
    - Create/Update Components.
    - Assemble in a View.
2. **Styling**:
    - Use Tailwind utility classes.
    - For custom colors, always use the `wa-*` or semantic classes defined in `tailwind.config.js`.
3. **Testing**:
    - Run `npm run test` to ensure no regressions.
