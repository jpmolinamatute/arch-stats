---
applyTo: "**/*.ts,**/*.js,**/*.vue"
---

# Frontend instructions

**Audience:** JavaScript/TypeScript developers working in `frontend/` directory

## Tech stack

- **Language:** TypeScript (strict)
- **Framework:** Vue 3 (Composition API, SFCs)
- **Build/Dev:** Vite ^6.3.x
- **Styling:** Tailwind CSS ^4.1.x (class-based dark mode enabled)
- **Quality:** ESLint + Prettier
- **API Types:** generated via `npm run generate:types` -> `frontend/src/types/types.generated.ts` (file must remain uncommitted / git ignored)
- **Runtime:** SPA at `http://localhost:5173`, proxied to backend at `http://localhost:8000/api/v0`

## Coding standards

- Use **Composition API** (`<script setup lang="ts">`) with explicit prop/event types.
- Keep components focused; prefer smaller components over large ones.
- **No implicit `any`**; define interfaces/types for props, emits, and API payloads.
- Re-use **generated API types** instead of redefining shapes.
- Centralize state in `frontend/src/state/` small composables in `frontend/src/composables/`.
- Prefer **fetch wrappers/clients** that return typed results rather than sprinkling raw `fetch` calls.
- Tailwind utility classes over ad-hoc CSS; avoid inline styles when possible.

### API Type Regeneration Workflow

Any backend schema or endpoint change (new field, rename, path) requires regenerating types:

1. Ensure backend dev server is running.
2. Run `npm run generate:types`.
3. Rebuild / restart Vite if types changed.
4. NEVER commit `types.generated.ts` (ensure it is listed in a .gitignore pattern such as `frontend/src/types/types.generated.ts`). Document in PR: "API types regenerated locally".

If the file ever appears in a diff, remove it and add the ignore rule before merging.

## File/Folder conventions

- Components: `PascalCase.vue` in `frontend/src/components/` (nest form components under `frontend/src/components/forms/`.
- Composables: `useThing.ts` in `frontend/src/composables/`.
- State: Pinia-like stores or simple modules in `frontend/src/state/`.
- Types: **import from** `frontend/src/types/types.generated.ts` for API schemas.

## Performance & DX

- Keep render trees shallow; memoize derived state using Vue's computed properties.
- Avoid reactivity leaks (do not spread reactive objects into plain objects).
- WebSockets: target pattern is to encapsulate connection logic inside a composable (e.g., `useShotsStream()`). CURRENT STATE: `ShotsTable.vue` holds inline WebSocket setup as a provisional implementation; refactor into a composable before adding additional WS consumers.
- Debounce user input that triggers network calls (≥150ms) to reduce chatter.
- Prefer `AbortController` to cancel in-flight fetches when switching views.
- Avoid large reactive objects in global state; store primitives/flat structures.

### WebSocket Messaging

Current prototype (`ShotsTable.vue`) receives raw `ShotsRead` JSON objects over `ws://.../api/v0/ws/shot` (no envelope). This will transition to a standardized envelope:

```json
{
  "type": "shot.created",
  "ts": "2025-08-20T12:34:56.789Z",
  "data": {
    /* domain payload (e.g., ShotsRead) */
  }
}
```

Rules (future state):

1. Unknown `type` values are ignored silently.
2. Heartbeat frames: `{ "type": "heartbeat", "ts": "..." }` update freshness only.
3. All domain payloads live under `data` and match generated OpenAPI types.

Action item: When backend adopts envelope, introduce a `useShotsStream()` composable that normalizes envelope formats during the transition window.

## Viewport support- Mobile-first:

The app must be fully responsive and work well in the following viewport widths:

- 360px
- 375px
- 390px
- 414px

## Do / Don't

### Do

- Use `defineProps<...>()` and `defineEmits<...>()` with explicit types.
- Validate inputs before hitting the API.
- Return `Promise<ResultType>` / typed promises from API helpers.
- Use the World Archery palette (base `wa.*` keys or semantic aliases: `primary`, `accent`, `success`, `warning`, `danger`, `info`, `neutral`) from `tailwind.config.js` instead of hard-coded hex values.

### Don't

- Commit `frontend/src/types/types.generated.ts` (add ignore if missing).
- Add heavy libraries without strong justification.
- Embed backend assumptions (e.g., SQL details) inside UI components.
- Keep long term WebSocket logic inline in components (move to composables during refactor).
- Directly manipulate DOM outside Vue unless using a properly isolated directive.

## References

- Architecture & cross-cutting rules: `.github/copilot-instructions.md`.
- For consistent endpoint usage patterns see existing composables (`useSession.ts`, `useTarget.ts`).
