# Sessions Page

## Feature Name

Sessions Management

## Description

Manage the creation, viewing, and closing of shooting sessions. Show details of the currently open session and a list of all past sessions, excluding the opened one.

## User Story

> As an archer, I want to create, view, and close practice sessions so that I can track my training progress over time.

## Acceptance Criteria

1. If no session is open:
   - Display a Session Creation Form with fields: `location`, `distance` (float, meters), `indoor` (bool).
   - Include a button to launch the Session Wizard.
2. If a session is open:
   - Display a Session Summary with: `start_time`, `location`, `distance`, `indoor`, and `targets` (just a placeholder).
   - Include a "Close session" button.
3. Display a Past Sessions Table with: `start_time`, `location`, `end_time`, `distance`, `indoor`, and `targets` (just a placeholder).
4. Exclude the open session from the Past Sessions Table.

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

- GET /api/v0/session
- GET /api/v0/session/open
- POST /api/v0/session
- GET /api/v0/session/{session_id}
- PATCH /api/v0/session/{session_id}
- DELETE /api/v0/session/{session_id}

## Backend files

- [backend/src/server/routers/sessions_router.py](https://github.com/jpmolinamatute/arch-stats/blob/main/backend/src/server/routers/sessions_router.py)

## Out of scope

- later on, we are going to manage targets (per session_id) in the Session section.
