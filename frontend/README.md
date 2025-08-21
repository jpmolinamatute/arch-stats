# Arch-Stats Frontend (aka WebUI)

- [Arch-Stats Frontend (aka WebUI)](#arch-stats-frontend-aka-webui)
    - [Overview](#overview)
    - [What This App Does](#what-this-app-does)
    - [What You'll Need](#what-youll-need)
    - [Setup: Step-by-Step](#setup-step-by-step)
        - [1. Install Dependencies](#1-install-dependencies)
        - [2. API Types](#2-api-types)
        - [3. Run the Dev Server](#3-run-the-dev-server)
    - [Project Structure](#project-structure)
    - [Tooling and Conventions](#tooling-and-conventions)
    - [Linting and Formatting](#linting-and-formatting)

## Overview

This guide is meant to help frontend developers to will walk you through:

- Setup their frontend development environment.
- Navigate the codebase.
- Use related tooling.

## What This App Does

The WebUI is a [Vue 3](https://vuejs.org/guide/introduction) + [TypeScript](https://www.typescriptlang.org/docs/) single-page app built with [Vite](https://vite.dev/guide/) and styled using [Tailwind CSS](https://tailwindcss.com/docs/installation/using-vite). It lets users:

- Register equipment (arrows, targets)
- Start, end and manage shooting sessions.
- View real-time shot data.
- Analyze data through charts and graphs.

It communicates with the API server via REST and WebSocket.

## What You'll Need

Before you start, make sure you have:

- npm version +11.4.2
- Linux Environment, Native Linux or WSL recommended. (macOS/Windows may need extra setup.)
- Arch-Stats backend server
- VS Code (optional), Project includes helpful workspace settings and tasks.

## Setup: Step-by-Step

### 1. Install Dependencies

Installs Vue, Vite, Tailwind, and other dependencies.

```bash
cd frontend
npm install
```

### 2. API Types

In order to generate API types from FastAPI's OpenAPI spec, the backend server must be running:

```bash
cd frontend
npm run generate:types
```

This pulls the OpenAPI schema from the API server and creates strongly typed interfaces in `frontend/src/types/types.generated.ts`. You need to run this command often in order to keep `frontend/src/types/types.generated.ts` up to date.

> **Note:** This file is git-ignored and should NOT be committed.

### 3. Run the Dev Server

Option 1: Using VS Code Tasks:

- Shortcut: `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux), then type and select `Tasks: Run Task`
- Choose **Start Vite Frontend**

Option 2: From the terminal:

```bash
cd frontend
npm run dev
```

The app will launch at <http://localhost:5173>. Edits will reload automatically.

## Project Structure

```text
frontend/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.app.json
└── src/
    ├── main.ts
    ├── App.vue
    ├── style.css
    ├── components/
    │   ├── WizardSession.vue
    │   ├── ShotsTable.vue
    │   └── forms/
    │       ├── NewArrow.vue
    │       ├── NewSession.vue
    │       └── CalibrateTarget.vue
    ├── components/widgets/Wizard.vue
    ├── composables/useSession.ts
    ├── state/
    │   ├── session.ts
    │   └── uiManagerStore.ts
    ├── types/types.generated.ts
```

## Tooling and Conventions

[Vite](https://vite.dev/guide/) is configured using [vite.config.ts](./vite.config.ts) file and it's responsible for:

- Fast reloads and development server.
- Proxy all `/api/v0/...` API calls to <http://localhost:8000>
- Build static assets into [backend/src/server/frontend/](../backend/src/server/frontend/). These assets will be served by the FastAPI backend in production.

```bash
cd frontend
npm run build
```

## Linting and Formatting

| Tool                                      | Command          |
| ----------------------------------------- | ---------------- |
| [ESLint](https://eslint.org/docs/latest/) | `npm run lint`   |
| [Prettier](https://prettier.io/docs/)     | `npm run format` |

> Linting and Auto-formatting are enabled if you use VS Code with [.vscode/settings.json](../.vscode/settings.json)
