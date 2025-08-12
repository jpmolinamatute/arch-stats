# Arrows Page

## Feature Name

Arrow Registration & Listing

## Description

Allow archers to register new arrows with metadata and view all registered arrows, showing whether each is programmed.

## User Story

> As an archer, I want to register my arrows and see their programmed status so I can track their performance individually.

## Acceptance Criteria

1. Display an Arrow Form with fields: `length`, `human_identifier`, `label_position`, `weight`, `diameter`, `spine`.
2. Display an Arrows Table with columns: `length`, `human_identifier`, `is_programmed` (indicator), `label_position`, `weight`, `diameter`, `spine`.
3. `is_programmed` must show a clear visual indicator (icon or tag).
4. Optionally allow launching the Wizard to program an arrow.

## UI Requirements

The UI for this feature must follow the global [UI & Look-and-Feel Requirements document](./00_UI_requirements.md), including:
Tailwind CSS with the extended World Archery color palette (wa-reflex, wa-pink, wa-yellow, wa-green, wa-red, wa-sky, wa-black).
Dark mode only theme using the specified palette and typography (Inter font).
Rounded corners for all elements (cards, panels, buttons, inputs).
Layout, table, form, button, and indicator styles exactly as defined in the global document.
Desktop-first responsiveness.
Any deviation from these standards must be documented and approved before implementation.

## Technical Requirements

- Must use: **Vue 3**, **TypeScript**, **Tailwind CSS**
- Integrate with these files:
  - [frontend/index.html](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/index.html)
  - [frontend/src/App.vue](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/src/App.vue)
  - [frontend/src/main.ts](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/src/main.ts)
  - [frontend/src/state/session.ts](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/src/state/session.ts)
  - [frontend/src/state/uiManagerStore.ts](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/src/state/uiManagerStore.ts)
  - frontend/src/types/types.generated.ts (this file will be provided separately)

## API references

- GET /api/v0/arrow/new_arrow_uuid
- GET /api/v0/arrow
- POST /api/v0/arrow
- GET /api/v0/arrow/{arrow_id}
- PATCH /api/v0/arrow/{arrow_id}
- DELETE /api/v0/arrow/{arrow_id}

## Backend files

- [backend/src/server/routers/arrows_router.py](https://github.com/jpmolinamatute/arch-stats/blob/main/backend/src/server/routers/arrows_router.py)
