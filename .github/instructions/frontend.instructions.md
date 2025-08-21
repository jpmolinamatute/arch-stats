# Frontend instructions

**Audience:** JavaScript/TypeScript developers working in [frontend/](../../frontend/)

---
applyTo: "**/*.ts,**/*.vue,**/*.js"
---

## Tech stack

- **Language:** TypeScript (strict)
- **Framework:** Vue 3 (Composition API, SFCs)
- **Build/Dev:** Vite
- **Styling:** Tailwind CSS (dark mode only for now)
- **Quality:** ESLint + Prettier
- **Types from API:** generated via `npm run generate:types` -> [frontend/src/types/types.generated.ts](../../frontend/src/types/types.generated.ts)
- **Runtime:** SPA at `http://localhost:5173`, proxied to backend at `http://localhost:8000/api/v0`

## Coding standards

- Use **Composition API** (`<script setup lang="ts">`) with explicit prop/event types.
- Keep components focused; prefer smaller components over large ones.
- **No implicit `any`**; define interfaces/types for props, emits, and API payloads.
- Re-use **generated API types** instead of redefining shapes.
- Centralize state in [frontend/src/state/](../../frontend/src/state/) or small composables in [frontend/src/composables/](../../frontend/src/composables/).
- Prefer **fetch wrappers/clients** that return typed results rather than sprinkling raw `fetch` calls.
- Tailwind utility classes over ad-hoc CSS; avoid inline styles when possible.

### API Type Regeneration Workflow

Any backend schema or endpoint change (new field, rename, path) requires regenerating types:
1. Ensure backend dev server is running.
2. Run `npm run generate:types`.
3. Rebuild / restart Vite if types changed.
4. NEVER commit `types.generated.ts` (ignored). Document in PR: "API types regenerated locally".

## File/Folder conventions

- Components: `PascalCase.vue` in frontend/src/components/ (nest form components under [frontend/src/components/forms/](../../frontend/src/components/forms/)).
- Composables: `useThing.ts` in frontend/src/composables/.
- State: Pinia-like stores or simple modules in [frontend/src/state/](../../frontend/src/state/).
- Types: **import from** [frontend/src/types/types.generated.ts](../../frontend/src/types/types.generated.ts) for API schemas.

## Performance & DX

- Keep render trees shallow; memoize derived state using Vue's computed properties.
- Avoid reactivity leaks (do not spread reactive objects into plain objects).
- For WebSockets, encapsulate connection logic in a composable (e.g., `useSession.ts`) and expose a small typed API.
- Debounce user input that triggers network calls (≥150ms) to reduce chatter.
- Prefer `AbortController` to cancel in-flight fetches when switching views.
- Avoid large reactive objects in global state; store primitives/flat structures.

### WebSocket Envelope (Future)

When adding WS support, standardize frames as JSON with minimal envelope:

```json
{
	"type": "shot.created", // event name
	"ts": "2025-08-20T12:34:56.789Z",
	"data": { /* domain payload */ }
}
```

Client silently ignores unknown `type` values; heartbeat frames (`type=heartbeat`) update connection freshness only.

## Do / Don't

### Do

- Use `defineProps<...>()` and `defineEmits<...>()` with explicit types.
- Validate inputs before hitting the API.
- Return `Promise<ResultType>` from API helpers.

### Don't

- Commit [frontend/src/types/types.generated.ts](../../frontend/src/types/types.generated.ts) (it's generated; ensure `.gitignore`).
- Add heavy libraries without strong justification.
- Embed backend assumptions in UI components.
- Directly manipulate DOM outside Vue unless using a properly isolated directive.

## References

- See [frontend/README.md](../../frontend/README.md) for structure and VS Code tasks.
 - Architecture & cross-cutting rules: `.github/copilot-instructions.md`.
 - For consistent endpoint usage patterns see existing composables (`useSession.ts`, `useTarget.ts`).
