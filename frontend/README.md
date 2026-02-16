# Arch Stats Frontend

## Description

**Arch Stats frontend** is a specialized application for tracking and analyzing archery statistics.
This frontend provides a modern, mobile-first interface for archers to log their sessions, visualize
their performance, and track their progress over time. It is built to be fast, responsive, and easy
to use on the range.

## Development Guidelines

**Audience:** Developers working in the `frontend/` directory.

## Architecture

The frontend is a Single Page Application (SPA) built with a modern stack:

- **Framework**: [Vue 3](https://vuejs.org/) (Composition API, SFCs)
- **Language**: [TypeScript](https://www.typescriptlang.org/) (Strict mode)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Styling**: [Tailwind CSS](https://tailwindcss.com/) (v4, class-based dark mode)
- **State Management**: Minimal global state; prefers Composables.
- **API Integration**: Generated TypeScript clients from OpenAPI specs.

It communicates with a Python backend (FastAPI/Uvicorn) served at `http://localhost:8000`.

## Golden Constraints (must-follow)

- **Composition API**: Use `<script setup lang="ts">` only.
- **Generated types**: Import API entities from `types.generated.ts`.
- **API calls in composables**: Components must not call `fetch()` directly.
- **Strict typing**: No implicit `any`; keep TypeScript strict.
- **Formatting**: ESLint + Prettier; keep diffs minimal and focused.
- **Build output path**: `npm run build` emits to `../backend/src/frontend/`.

## Prerequisites

Before you start, ensure your environment is ready:

- **Node.js**: v25.2.1+
- **npm**: v11.6.4+
- **Backend**: The backend service must be running for full functionality.

## Setup

Follow these steps to get from zero to "Hello World":

1. **Install Dependencies**:

    ```bash
    npm install
    ```

2. **Generate API Types**:
    Ensure the backend is running, then generate the TypeScript client:

    ```bash
    npm run generate:types
    ```

    > [!IMPORTANT]
    > This step is crucial to ensure your frontend types match the latest backend schema. Run this
    > whenever the backend API changes.

3. **Start Development Server**:

    ```bash
    npm run dev
    ```

    The app will be available at `http://localhost:5173`.

4. **Start Backend (Required for API)**:
    Open a new terminal and run:

    ```bash
    ./scripts/start_uvicorn.bash
    ```

## Quickstart (VS Code tasks)

If you use VS Code, the workspace defines helpful tasks:

- Start frontend dev server: Run task `Start Vite Server`.
- Start backend API: Run task `Start Uvicorn Server`.
- Start/stop database: Run tasks `Start Docker Compose` and
  `Stop Docker Compose` (or `Stop & Remove Volumes`).

## Scripts

Here are the common commands you'll need:

| Command | Description |
| :--- | :--- |
| `npm run dev` | Start the Vite development server. |
| `npm run build` | Type-check and build the production bundle. |
| `npm run preview` | Preview the production build locally. |
| `npm run lint` | Lint and fix code using ESLint. |
| `npm run test` | Run unit tests with Vitest. |
| `npm run generate:types` | Regenerate API types from the running backend. |

## Git Hooks & Safety Net

To prevent committing broken code, "activate" the git pre-commit hook by creating a symlink to the
linting script:

```bash
ln -sfr ./scripts/linting.bash ./.git/hooks/pre-commit
```

This ensures that all linters and tests pass before a commit is allowed.

### Pull Requests

After pushing your changes, use the helper script to create a Pull Request:

```bash
./scripts/create_pr.bash
```

> [!IMPORTANT]
> Workflows are triggered selectively based on the files changed (e.g., frontend changes do not
> trigger backend tests). All **triggered** workflows must pass for a PR to be mergeable.

## Structure

The codebase is organized by function:

- **`frontend/src/api/`**: API client configuration.
- **`frontend/src/components/`**: Vue Single File Components (SFCs). Kept small and presentation-focused.
- **`frontend/src/composables/`**: Reusable logic and state management. **All HTTP requests live here.**
- **`frontend/src/router/`**: Application routing configuration.
- **`frontend/src/types/`**: Generated TypeScript types (`types.generated.ts`).
- **`frontend/src/assets/`**: Static assets (images, icons, fonts).
- **`frontend/tests/`**: Unit tests using Vitest.

### Frontend contribution pattern

When adding a new feature or integrating a new API endpoint:

1. Ensure backend is running. Generate types with `npm run generate:types`.
2. Add a composable in `src/composables/` wrapping the API call and returning typed data.
3. Connect components to the composable; keep components presentational and small.
4. Add tests in `frontend/tests/` for components and composables.
5. Run `npm run lint` and ensure type checks pass via `npm run build`.

### Style Guidelines

> [!IMPORTANT]
> Follow the [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html) for consistency.

### Coding Standards

1. **Composition API**: Use `<script setup lang="ts">`.
2. **Explicit Types**: No implicit `any`. Define types for all props, emits, and data.
3. **Tailwind**: Use utility classes over inline styles.
4. **Type Sync**: Import entities from `types.generated.ts`.
5. **Logic Separation**: Keep complex logic in `composables/`.

### Contribution Pattern

When adding a new feature or integrating a new API endpoint:

1. Ensure backend is running. Generate types with `npm run generate:types`.
2. Add a composable in `src/composables/` wrapping the API call and returning typed data.
3. Connect components to the composable; keep components presentational and small.
4. Add tests in `frontend/tests/` for components and composables.
5. Run `npm run lint` and ensure type checks pass via `npm run build`.

### Mobile-First Design

This app is primarily used on phones.

- **Viewport**: Prioritize 360–414px widths.
- **Layout**: Default to `w-full` and `flex-col`. Use `md:`/`lg:` for larger screens.
- **Nesting**: Limit depth to **3 levels** of layout containers.

### API Access Rules

- **No Direct Fetch**: Components must not call `fetch()` directly.
- **Composables**: Wrap all API calls in composables returning typed Promises.
- **Type Generation**: Run `npm run generate:types` when the backend schema changes.

### Code Quality

All code must pass linting:

```bash
npm run lint
```

We use **ESLint** with the **Antfu** configuration.

### Performance

- Debounce inputs (≥150ms).
- Use `AbortController` for cancellable requests.
- Keep render trees shallow.

### Development Proxy

In development (`npm run dev`), Vite serves the frontend at `http://localhost:5173`. API requests to
`/api/v0` are proxied to the backend at `http://localhost:8000` to avoid CORS issues.

**Configuration (`vite.config.ts`):**

```typescript
server: {
  proxy: {
    '/api/v0': {
      target: 'http://localhost:8000',
      changeOrigin: true,
      ws: true,
    },
  },
}
```

### Production Build

In production, the frontend is built into static files and served directly by the backend (Uvicorn).

1. **Build the Frontend**:

    ```bash
    npm run build
    ```

    This compiles the app to `../backend/src/frontend/`.

2. **Serve with Uvicorn**:
    When you start the backend:

    ```bash
    ./scripts/start_uvicorn.bash
    ```

    Uvicorn serves the API at `/api/v0` AND the static frontend files at the root `/`.

## Troubleshooting

- Types missing or out of date: Ensure backend is running and run
  `npm run generate:types`.
- Proxy not working: Check `vite.config.ts` proxy to `http://localhost:8000` and ports.
- CORS errors in dev: Use the dev proxy (`/api/v0`), not absolute backend URLs.
- Node/npm issues: Ensure versions match the prerequisites listed above.
- Build path mismatch: Production build must emit to `../backend/src/frontend/`.

### Do / Don't

| ✅ Do | ❌ Don't |
| :--- | :--- |
| Use `defineProps` & `defineEmits` with types. | Add large libs without good reason. |
| Validate inputs before API calls. | Hard-code backend URLs/assumptions. |
| Use Tailwind semantic colors (`primary`, `warning`). | Avoid complex WS logic in components. |

## References

- **Backend**: [backend/README.md](../backend/README.md)
- **Scripts**: [scripts/README.md](../scripts/README.md)
