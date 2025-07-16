# Arch-Stats WebUI

Welcome to the Arch-Stats WebUI — the frontend interface for tracking and visualizing archery performance.

This guide will walk you through setup, structure, tooling, and development tips for working on the frontend.

## 🚀 What This App Does

The WebUI is a [Vue 3](https://vuejs.org/guide/introduction) + [TypeScript](https://www.typescriptlang.org/docs/) single-page app built with [Vite](https://vite.dev/guide/) and styled using [Tailwind CSS](https://tailwindcss.com/docs/installation/using-vite). It lets users:

- Register equipment (arrows, targets)
- Start and manage shooting sessions
- View real-time shot data

It communicates with the API server via REST and WebSocket.

## 🧰 What You'll Need

Before you start, make sure you have:

- **npm** version +11.4.2
- **Linux Environment** – Native Linux or WSL recommended. (macOS/Windows may need extra setup.)
- **VS Code (optional)** – Project includes helpful workspace settings and tasks.

## 🛠️ Setup: Step-by-Step

### 1. Install Dependencies

```bash
cd frontend
npm install
```

This installs Vue, Vite, Tailwind, and other dependencies.

### 2. API Types

In order to generate API types from FastAPI’s OpenAPI spec, the backend must be running:

```bash
cd frontend
npm run generate:types
```

This pulls the OpenAPI schema from the API server and creates strongly typed interfaces in `src/types/types.generated.ts`. You need to run this command often in order to keep `src/types/types.generated.ts` up to date.

> **Note:** This file is git-ignored and should not be committed.

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

## 📁 Project Structure

```text
frontend/
├── index.html
├── package.json
├── vite.config.ts          # Vite config (API proxy, build output)
├── tsconfig.app.json       # TS config
└── src/
    ├── main.ts
    ├── App.vue
    ├── style.css           # Tailwind/global styles
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

## ⚙️ Tooling & Conventions

[Vite](https://vite.dev/guide/) is configured using [vite.config.ts](./vite.config.ts) file and it's responsible for:

- Fast reloads and development server.
- Proxy all `/api/v0/...` API calls to <http://localhost:8000>
- Build static assets into [backend/src/server/frontend/](../backend/src/server/frontend/). These assets will be served by the FastAPI backend in production.

```bash
cd frontend
npm run build
```

## 🧹 Linting & Formatting

| Tool                                      | Command          |
| ----------------------------------------- | ---------------- |
| [ESLint](https://eslint.org/docs/latest/) | `npm run lint`   |
| [Prettier](https://prettier.io/docs/)     | `npm run format` |

> Linting and Auto-formatting are enabled if you use VS Code with [.vscode/settings.json](../.vscode/settings.json)

## ✅ You're Ready

You’ve got the frontend running. You can now:

- Use the interface to register equipment and sessions
- View real-time archery data
- Add features or improve UX/UI
