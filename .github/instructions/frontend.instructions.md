---
applyTo: "**/*.ts,**/*.js,**/*.vue"
---

# Frontend Development Instructions

**Audience:** Developers working in the `frontend/` directory using Vue 3 + TypeScript.

## Tech Stack

- **Language:** TypeScript (strict)
- **Framework:** Vue 3 (Composition API, SFCs)
- **Build/Dev:** Vite ^6.3.x
- **Styling:** Tailwind CSS ^4.1.x (class-based dark mode)
- **Quality:** ESLint + Prettier
- **API Types:** generated via `npm run generate:types` → `frontend/src/types/types.generated.ts`
- **Runtime:** SPA served at `http://localhost:5173`, proxied to backend `http://localhost:8000/api/v0`

## Project Organization

The project has two main functional directories:

- **`frontend/src/components/`** — Vue SFCs in `PascalCase.vue`.
  Keep components small, cohesive, and presentation-focused.
- **`frontend/src/composables/`** — Reusable logic modules (`useThing.ts`).
  Encapsulate all **HTTP requests** here.

### API Access Rules

- All `fetch()` or HTTP client logic must live inside composables.  
  Components **must never** call `fetch()` directly.
- Wrap API calls in well-typed functions returning `Promise<ResultType>`.
- Avoid duplication — extract shared logic into reusable composables  
  (e.g., `useApiFetch()`, `useSessionApi()`).

### Type Synchronization

- Always import types from `frontend/src/types/types.generated.ts`.
- Regenerate types whenever backend schemas change:

```bash
npm run generate:types
```

- Never commit `types.generated.ts`; ensure `.gitignore` includes it.
- Treat manually defined API types as temporary placeholders only.

## Coding Standards

1. Use the **Composition API** (`<script setup lang="ts">`) with explicit prop/event types.
2. No implicit `any`. All data, props, and emits must have defined types.
3. Prefer smaller components; one logical concern per file.
4. Keep reactivity localized; do not spread reactive objects.
5. Use Tailwind utility classes instead of inline styles.
6. Import from `types.generated.ts` whenever referring to backend entities.
7. Manage global or shared state through lightweight composables in `src/composables/`.

## Mobile-First Design

This webapp is **mobile-first**; over 95% of traffic is expected from phones.

Copilot must:

- Prioritize narrow viewports (360–414 px).
- Default to vertical stacking and full-width layouts (`w-full`, `flex-col`).
- Scale up using responsive utilities (`md:`, `lg:`) rather than fixed widths.
- Avoid desktop-first or pixel-perfect assumptions.
- Keep layouts **flat** and **lightweight**.

### Nesting Rule

Limit template depth: **no more than three nested layout containers**
(e.g., `<div><div><div>...</div></div></div>`).
Deep nesting reduces readability and complicates debugging.
Use grid/flex utilities or semantic grouping instead of wrapper divs.

## Code Quality

All frontend code must pass linting and formatting before commit.

### Tools

- **ESLint** — code quality and rule enforcement
- **Prettier** — formatting

You can run these tools individually:

```bash
npm run lint
npm run format
```

or together:

```bash
./scripts/linting.bash --lint-frontend
```

Copilot must emit code that conforms to ESLint rules and Prettier formatting automatically.

## Performance & Developer Experience

- Keep render trees shallow; prefer `computed()` over large reactive objects.
- Debounce network-triggering inputs (≥150 ms).
- Use `AbortController` to cancel in-flight requests.
- Encapsulate WebSocket or event-stream logic inside composables (e.g., `useShotsStream()`).
- Memoize derived state where possible.
- Prefer primitive/flat structures in global state.

### WebSocket Messaging (current prototype)

The active implementation in `ShotsTable.vue` receives raw JSON (`ShotsRead`) over
`ws://.../api/v0/ws/shot`. This will evolve to a standardized envelope:

```json
{
  "type": "shot.created",
  "ts": "2025-08-20T12:34:56.789Z",
  "data": {
    /* domain payload */
  }
}
```

Rules for future state:

1. Unknown `type` values are ignored silently.
2. Heartbeats (`type: "heartbeat"`) update freshness.
3. All domain payloads live under `data` and match generated OpenAPI types.

When the backend adopts this envelope, refactor into `useShotsStream()`.

## Do / Don’t

### ✅ Do

- Use `defineProps<...>()` and `defineEmits<...>()` with explicit types.
- Validate user inputs before calling APIs.
- Return typed promises from API helpers.
- Use Tailwind semantic color keys (`primary`, `accent`, `warning`, etc.) from `tailwind.config.js`.

### ❌ Don’t

- Commit `frontend/src/types/types.generated.ts`.
- Add large external libraries without clear ROI.
- Hard-code backend assumptions in UI logic.
- Inline long-term WebSocket code inside components.
- Directly manipulate the DOM outside Vue or proper directives.

## References

- Architecture & cross-cutting rules: `.github/copilot-instructions.md`
- Backend guidelines: `.github/instructions/backend.instructions.md`
- Existing composables for reference: `useSession.ts`, `useTarget.ts`

## Summary

Copilot must:

- Keep API calls inside `src/composables/`.
- Use and regenerate `types.generated.ts` frequently.
- Generate simple, mobile-first layouts (≤3 nested wrappers).
- Follow ESLint + Prettier formatting.
- Use Composition API and strict typing.
- Favor reusability and clarity over nesting and complexity.
