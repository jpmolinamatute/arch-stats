# Shots Page

## Feature Name

Live Shot Tracking

## Description

Display live shot data in a table with real-time updates via WebSocket.

## User Story

> As an archer, I want to see my shots appear in real time so I can adjust my performance during practice.

## Acceptance Criteria

1. Display a Shots Table with columns from the backend shot schema.
2. Connect to the WebSocket feed when the Shots page is mounted.
3. Disconnect from the WebSocket when navigating away from the page or opening the Wizard.
4. Optionally show a connection status indicator.

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
- frontend/src/types/types.generated.ts (this file will be provided separately)

## API references

- GET /api/v0/shot
- DELETE /api/v0/shot/{shot_id}
- WEBSOCKET /api/v0/ws/shot

## Backend files

[backend/src/server/routers/shots_router.py](https://github.com/jpmolinamatute/arch-stats/blob/main/backend/src/server/routers/shots_router.py)
[backend/src/server/routers/websocket.py](https://github.com/jpmolinamatute/arch-stats/blob/main/backend/src/server/routers/websocket.py)
