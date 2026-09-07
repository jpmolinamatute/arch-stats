# Product Requirements Document (PRD) & Technical Specification

## Historical Analytics & Visual Charting Feature

- **Target Route:** `/app/analytics`
- **Feature Owner:** Analytics & Frontend / Backend Core
- **Status:** Approved / Ready for Implementation Plan
- **Date:** 2026-09-06
- **Storage Location:** `docs/analytics_feature/prd.md`

## 1. Executive Summary & Goals

### 1.1 Problem Statement

In archery, tracking performance over time is essential for equipment tuning, form consistency, and
tournament preparation. Currently, the platform collects real-time shot coordinates and scores
during live shooting sessions, but lacks a centralized, historical analytics hub where archers can
visualize long-term progression, analyze target dispersion, and query their stats across venues,
dates, and bowstyles.

### 1.2 Core Value Proposition

Provide archers with an intuitive, visually stunning, and responsive
**Historical AnalyticsDashboard** (`/app/analytics`) that answers three critical questions at a
glance:

1. **"Am I improving?"** (Scoring progression trendline and moving averages).
2. **"Where do my arrows cluster?"** (Target face dispersion, center-of-mass drift, and grouping
   radius).
3. **"What is my scoring consistency?"** (Hit distribution across target rings).

### 1.3 Key Architectural Decisions

- **PostgreSQL-First Computation:** All complex data manipulation, math (center-of-mass,
  groupradius, grouping, histograms) is encapsulated in PostgreSQL 17 stored functions via Goose
  migrations. The Go chi backend server remains thin, executing simple `SELECT * FROM get_...()`
  queries.
- **Granular Micro-Endpoints:** Independent REST endpoints (`/filters`, `/kpis`, `/trends`,
  `/distribution`, `/dispersion`, `/session/{id}`) allow the frontend to load widgets concurrently
  with non-blocking skeleton loaders.
- **Zero-Latency Auth Caching:** `archer_id` is cached synchronously in `localStorage` upon login,
  allowing the dashboard to bootstrap and fire analytics queries immediately on mount without waiting
  for `/auth/me`.
- **Lightweight Responsive Charting:** Chart.js via `vue-chartjs` powers time-series trends and ring
  accuracy, while the existing SVG engine powers target dispersion.

## 2. User Experience & Design

### 2.1 Information Architecture & Layout (Approach 1: Segmented Analytics Hub)

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                    AppHeader.vue (Adds "Analytics" link)                 │
└──────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    AnalyticsDashboard.vue (/app/analytics)               │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 1. AnalyticsFilterBar.vue (Sticky Top Bar)                         │  │
│  │    [Date Range: Last 90D ▾] [Location ▾] [Bowstyle ▾] [Indoor/Out ▾]│ │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 2. AnalyticsKPIDeck.vue (4 Hero Cards with Skeleton Loaders)       │  │
│  │    [ Avg Arrow Score ] [ Total Arrows ] [ Personal Best ] [ X-Rate ]│ │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 3. ScoringTrendChart.vue (Chart.js via vue-chartjs)                │  │
│  │    • Segmented toggle: [Average Arrow Score (default) | Total]     │  │
│  │    • Spline curve + 5-session moving average                       │  │
│  │    • Click point ➔ opens Session Detail Drawer                     │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌─────────────────────────────────┬──────────────────────────────────┐  │
│  │ 4. ScoreDistributionChart.vue   │ 5. TargetDispersionView.vue      │  │
│  │    (Ring Accuracy Bar / Donut)  │    (SVG Target Face Grouping)    │  │
│  │    • Standard Ring Colors:      │    • Group tabs: [Indoor 40cm]   │  │
│  │      Gold (X,10,9), Red (8,7),  │    • Overlays Center-of-Mass (cx)│  │
│  │      Blue (6,5), Black, White   │    • Draws Group Radius circle   │  │
│  └─────────────────────────────────┴──────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 6. SessionDetailDrawer.vue (Slide-Out Side Panel / Modal)          │  │
│  │    • Opened upon clicking a session node on the trend chart        │  │
│  │    • Displays single-session scorecard, slot details, & target SVG │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Component Specifications

#### 2.2.1 `AnalyticsFilterBar.vue` (Always-Visible Top Filter Bar)

- **Positioning:** Sticky below `AppHeader.vue`.
- **Controls:**
  1. **Date Range Dropdown:**
     - Presets: "Last 7 Days", "Last 30 Days", **"Last 90 Days" (Default)**, "This Season",
       "All Time", and "Custom Range" (revealing start/end date pickers).
  2. **Location Dropdown:**
     - Options: "All Locations" (Default) plus dynamically loaded distinct locations where the
       archer has shot.
  3. **Bowstyle Dropdown:**
     - Options: "All Bowstyles" (Default), "Recurve", "Compound", "Barebow", "Longbow".
  4. **Environment Dropdown:**
     - Options: "All Environments" (Default), "Indoor Only", "Outdoor Only".
- **URL Synchronization:** All active filters sync to URL query parameters
  (`?range=90d&bowstyle=recurve&location=Club...`) allowing direct bookmarking and reload
  resilience.

#### 2.2.2 `AnalyticsKPIDeck.vue` (High-Impact Hero Cards)

- **Card 1: Arrow Average:** Overall mean score per arrow (e.g. `9.24 / 10`).
- **Card 2: Total Arrows Shot:** Total volume of arrows logged in the period (e.g. `1,420`).
- **Card 3: Personal Best:** Highest single-session score recorded in the period (e.g. `294 / 300`).
- **Card 4: X-Ring / 10 Rate:** Percentage of arrows in the 10/X ring (e.g. `46.8% (665 Xs)`).
- **States:** Shimmer skeleton animation while queries resolve; `--` fallback if no data.

#### 2.2.3 `ScoringTrendChart.vue` (Progression Timeline)

- **Engine:** Chart.js with `vue-chartjs`.
- **Controls:** Segmented toggle in card header:
    - `Average Arrow Score` (Default, 0.0 to 10.0 scale, enabling fair comparisons across sessions
      of varying length).
    - `Total Session Score` (Sum of points scored).
- **Features:**
    - Smooth spline line connecting chronologically ordered sessions.
    - Optional 5-session rolling moving average trendline to smooth out variability.
    - Tooltip: Date, Location, Total Arrows, Arrow Average, and Session Total.
    - **Interaction:** Clicking any session point fires `select-session(sessionId)` to trigger the
      drill-down drawer.

#### 2.2.4 `ScoreDistributionChart.vue` (Ring Accuracy)

- **Engine:** Chart.js horizontal bar or donut chart.
- **Visuals:** Color-coded to World Archery target standards:
    - **Gold:** X and 10 (and 9)
    - **Red:** 8 and 7
    - **Blue:** 6 and 5
    - **Black:** 4 and 3
    - **White:** 2 and 1
    - **Gray:** M (Miss / 0)
- **Metrics:** Hit count and exact percentage for each ring.

#### 2.2.5 `TargetDispersionView.vue` (Group Geometry & Heatmap)

- **Engine:** Extended SVG target face renderer based on [Face.vue](file:///home/juanpa/Projects/arch-stats/frontend/src/components/Face.vue).
- **Group Selector:** Segmented pill tabs grouping shots by target face and environment (e.g.,
  `[Indoor WA 40cm (320 shots)]`, `[Outdoor WA 122cm (180 shots)]`).
- **Geometric Overlay:**
    - **Plotted Shots:** Semi-transparent circles showing all shot impacts.
    - **Center of Mass (Crosshair):** Plotted at calculated `(center_x, center_y)` to visualize
      horizontal/vertical group drift.
    - **Group Radius (Circle):** Outlined circle centered on the center-of-mass encompassing the
      group cluster spread.
    - **Spread Metric Badge:** Shows group diameter in SVG units/mm (e.g. `Group Spread: 54mm`).

#### 2.2.6 `SessionDetailDrawer.vue` (Interactive Drill-Down)

- **UX Pattern:** Slide-over right drawer (desktop) / bottom sheet (mobile) with dimmed backdrop.
- **Content:**
    - Header: Session Date, Venue, Environment, Bowstyle, Total Arrows, Final Score.
    - Target Face: Mini SVG showing the exact shots from that individual session.
    - End-by-End Scorecard: Table showing arrow scores per end (e.g. End 1: 10, 9, 9 = 28).
- **Dismissal:** ESC key, top-right "X" button, or tapping the backdrop.

## 3. System Architecture & Database Layer

### 3.1 Architecture Overview

```text
┌────────────────────────────────────────────────────────┐
│                   Go Backend Server                    │
│      Thin HTTP Handlers ➔ ReportingRepo (One-Liners)   │
│         "SELECT * FROM get_archer_kpis(...)"           │
└───────────────────────────┬────────────────────────────┘
                            │ (Parameters: ID, Filters)
                            ▼
┌────────────────────────────────────────────────────────┐
│                   PostgreSQL 17                        │
│                                                        │
│  ├── get_archer_analytics_filters(p_archer_id)          │
│  ├── get_archer_kpis(p_archer_id, p_from, p_to, ...)   │
│  ├── get_archer_trends(p_archer_id, p_from, ...)       │
│  ├── get_archer_distribution(p_archer_id, ...)         │
│  ├── get_archer_dispersion(p_archer_id, ...)           │
│  │     • Computes Center of Mass (AVG(x), AVG(y))      │
│  │     • Computes Group Radius SQRT((x-cx)² + (y-cy)²) │
│  │     • Groups by face_type & is_indoor via json_agg  │
│  └── get_archer_session_detail(p_archer_id, p_session) │
└────────────────────────────────────────────────────────┘
```

### 3.2 Database Migrations (Goose)

A new migration `backend/migrations/007_2026-09-06_analytics_functions.sql` defines:

#### 3.2.1 Indexes

```sql
CREATE INDEX IF NOT EXISTS idx_slot_archer_session ON slot(archer_id, session_id);
CREATE INDEX IF NOT EXISTS idx_session_owner_created ON session(owner_archer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shot_slot_score ON shot(slot_id, score, is_x);
```

#### 3.2.2 PostgreSQL Stored Functions

1. **`get_archer_analytics_filters(p_archer_id UUID)`**
   - Extracts:
     - Distinct locations: `ARRAY_AGG(DISTINCT session.session_location)`
     - Distinct bowstyles: `ARRAY_AGG(DISTINCT slot.bowstyle)`
     - Environment booleans: `BOOL_OR(session.is_indoor)`, `BOOL_OR(NOT session.is_indoor)`
     - Date limits: `MIN(session.created_at)`, `MAX(session.created_at)`
   - Returns a single JSON object.
2. **`get_archer_kpis(p_archer_id UUID, p_from TIMESTAMPTZ, p_to TIMESTAMPTZ, p_location TEXT,
   p_bowstyle BOWSTYLE_TYPE, p_is_indoor BOOLEAN)`**
   - Returns:
     - `total_sessions`: `COUNT(DISTINCT session.session_id)`
     - `total_shots`: `COUNT(shot.shot_id)`
     - `average_score`: `COALESCE(ROUND(AVG(shot.score)::numeric, 2), 0.0)`
     - `personal_best`: `COALESCE(MAX(session_scores.total_score), 0)`
     - `total_x_count`: `COUNT(CASE WHEN shot.is_x THEN 1 END)`
     - `x_rate`: `ROUND((COUNT(CASE WHEN shot.is_x THEN 1 END)::numeric / NULLIF(COUNT(shot.shot_id)
       , 0)) * 100, 1)`
3. **`get_archer_trends(p_archer_id UUID, p_from TIMESTAMPTZ, p_to TIMESTAMPTZ, p_location TEXT,
   p_bowstyle BOWSTYLE_TYPE, p_is_indoor BOOLEAN)`**
   - Aggregates by `session.session_id` and orders by `session.created_at ASC`.
   - Returns table:
     - `session_id UUID`
     - `date TIMESTAMPTZ`
     - `session_location VARCHAR(255)`
     - `is_indoor BOOLEAN`
     - `bowstyle BOWSTYLE_TYPE`
     - `total_shots INTEGER`
     - `total_score INTEGER`
     - `average_score DOUBLE PRECISION`
     - `x_count INTEGER`
4. **`get_archer_distribution(p_archer_id UUID, p_from TIMESTAMPTZ, p_to TIMESTAMPTZ, p_location
   TEXT, p_bowstyle BOWSTYLE_TYPE, p_is_indoor BOOLEAN)`**
   - Groups by `shot.score` and `shot.is_x`.
   - Returns a single JSON object containing exact shot counts and calculated percentages for values
     `"0"` through `"10"` and `"X"`.
5. **`get_archer_dispersion(p_archer_id UUID, p_from TIMESTAMPTZ, p_to TIMESTAMPTZ, p_location TEXT,
   p_bowstyle BOWSTYLE_TYPE, p_is_indoor BOOLEAN)`**
   - Filters shots where `x IS NOT NULL AND y IS NOT NULL`.
   - Groups by `slot.face_type` and `session.is_indoor`.
   - Encapsulates geometry math in SQL:
     - `center_x = AVG(shot.x)`
     - `center_y = AVG(shot.y)`
     - `group_radius = COALESCE(MAX(SQRT(POWER(shot.x - sub.cx, 2) + POWER(shot.y - sub.cy, 2))),
       0.0)`
     - `shots = json_agg(json_build_object('x', shot.x, 'y', shot.y, 'score', shot.score, 'is_x',
       shot.is_x))`
   - Returns a row per `(face_type, is_indoor)` group.
6. **`get_archer_session_detail(p_archer_id UUID, p_session_id UUID)`**
   - Returns a single JSON object representing the full session: metadata, slot settings, distance,
     lane, ends, shots array, and total score for the drill-down drawer.

## 4. Backend API Specification

All endpoints are mounted on Chi router under `/api/v1/analytics/archer/{archer_id}` and protected
by `requireOwnership(w, r)`.

### 4.1 Filter Query Parameters

Endpoints `GET /kpis`, `GET /trends`, `GET /distribution`, and `GET /dispersion` accept standardized
query parameters:

- `from` (string, ISO 8601 timestamp, optional)
- `to` (string, ISO 8601 timestamp, optional)
- `location` (string, optional)
- `bowstyle` (string enum: `recurve`, `compound`, `barebow`, `longbow`, optional)
- `is_indoor` (boolean, optional)

### 4.2 Endpoint Definitions

#### 4.2.1 `GET /api/v1/analytics/archer/{archer_id}/filters`

- **Description:** Returns metadata to populate filter controls.
- **Success 200 Response:**

```json
{
  "locations": ["Indoor Range 1", "City Field"],
  "bowstyles": ["recurve", "barebow"],
  "has_indoor": true,
  "has_outdoor": true,
  "min_date": "2025-10-01T10:00:00Z",
  "max_date": "2026-09-01T14:30:00Z"
}
```

#### 4.2.2 `GET /api/v1/analytics/archer/{archer_id}/kpis`

- **Description:** Returns high-level career/filter metrics.
- **Success 200 Response:**

```json
{
  "total_sessions": 24,
  "total_shots": 1420,
  "average_score": 9.24,
  "personal_best": 294,
  "total_x_count": 665,
  "x_rate": 46.8
}
```

#### 4.2.3 `GET /api/v1/analytics/archer/{archer_id}/trends`

- **Description:** Returns chronological session trendline points.
- **Success 200 Response:**

```json
[
  {
    "session_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "date": "2026-06-15T18:00:00Z",
    "session_location": "Indoor Range 1",
    "is_indoor": true,
    "bowstyle": "recurve",
    "total_shots": 60,
    "total_score": 562,
    "average_score": 9.37,
    "x_count": 28
  }
]
```

#### 4.2.4 `GET /api/v1/analytics/archer/{archer_id}/distribution`

- **Description:** Returns frequency counts and percentages of scores.
- **Success 200 Response:**

```json
{
  "total_scored_shots": 1400,
  "counts": {
    "X": 665,
    "10": 280,
    "9": 320,
    "8": 100,
    "7": 25,
    "6": 8,
    "5": 2,
    "4": 0,
    "3": 0,
    "2": 0,
    "1": 0,
    "0": 0
  },
  "percentages": {
    "X": 47.5,
    "10": 20.0,
    "9": 22.9,
    "8": 7.1,
    "7": 1.8,
    "6": 0.6,
    "5": 0.1,
    "4": 0.0,
    "3": 0.0,
    "2": 0.0,
    "1": 0.0,
    "0": 0.0
  }
}
```

#### 4.2.5 `GET /api/v1/analytics/archer/{archer_id}/dispersion`

- **Description:** Returns target face dispersion groups with geometry metrics.
- **Success 200 Response:**

```json
[
  {
    "face_type": "wa_40cm_full",
    "is_indoor": true,
    "total_shots": 320,
    "center_x": 150.2,
    "center_y": 149.8,
    "group_radius": 24.5,
    "shots": [
      { "x": 150.1, "y": 148.9, "score": 10, "is_x": true },
      { "x": 147.5, "y": 152.0, "score": 9, "is_x": false }
    ]
  }
]
```

#### 4.2.6 `GET /api/v1/analytics/archer/{archer_id}/session/{session_id}`

- **Description:** Returns denormalized session detail for drill-down drawer.
- **Success 200 Response:**

```json
{
  "session_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "created_at": "2026-06-15T18:00:00Z",
  "session_location": "Indoor Range 1",
  "is_indoor": true,
  "bowstyle": "recurve",
  "distance": 18,
  "face_type": "wa_40cm_full",
  "total_score": 562,
  "total_shots": 60,
  "average_score": 9.37,
  "shots": [
    { "shot_id": "...", "score": 10, "is_x": true, "x": 150.1, "y": 148.9, "created_at": "..." }
  ]
}
```

## 5. Frontend Architecture & State Management

### 5.1 Auth Caching Optimization

- In `frontend/src/composables/useAuth.ts`:
    - Upon successful login (or on `/auth/me` return), `archer_id` is persisted to
      `localStorage.setItem('arch-stats:archer_id', archerId)`.
    - On logout, `localStorage.removeItem('arch-stats:archer_id')`.
- `useAnalytics.ts` reads `archer_id` synchronously from `localStorage`, immediately firing analytics
  fetches on component mount with 0ms latency.

### 5.2 Composable `useAnalytics.ts`

- **State:**
    - `filters`: `{ range: '90d', startDate: null, endDate: null, location: null, bowstyle: null,
      isIndoor: null }`
    - `filterOptions`: `{ locations: [], bowstyles: [], hasIndoor: true, hasOutdoor: true }`
    - `kpis`: Ref<ArcherKPIReport | null>
    - `trends`: Ref<SessionTrendPoint[]>
    - `distribution`: Ref<ScoreDistribution | null>
    - `dispersion`: Ref<TargetFaceDispersionGroup[]>
    - `selectedSession`: Ref<SessionDetailReport | null>
    - Loading states: `loadingFilters`, `kpisLoading`, `trendsLoading`, `distributionLoading`,
      `dispersionLoading`, `sessionDetailLoading`.
- **Actions:**
    - `initFromURL()`: Parses query string into `filters`.
    - `updateFilters(newFilters)`: Pushes query to `router.replace()`, aborts previous requests via
      `AbortController`, and refetches micro-endpoints in parallel.
    - `selectSession(sessionId)`: Fetches `/session/{id}` and opens drawer.

### 5.3 Empty & Edge State Handling

- **New Archer:** When `total_sessions == 0` across all time, renders welcoming onboarding card with
  CTA **"Start Your First Session"** routing to `/app`.
- **No Filter Match:** When active filters yield 0 sessions, renders an inline banner:
  **"No sessions match this filter"** with a **"Reset Filters"** button.
- **Unscored / Practice Sessions:** Handled seamlessly by separating `total_shots` from scored shot
  calculations; averages ignore null scores.
- **Single-Shot Sessions:** Dispersion sets `group_radius = 0.0` and center-of-mass at the single
  shot coordinate to prevent math exceptions.

## 6. Testing & Verification Plan

### 6.1 Database Verification

- Execute Goose migration `007_2026-09-06_analytics_functions.sql` up and down to verify syntax and
  idempotency.
- Verify PL/pgSQL boundary conditions:
    - Division by zero on empty archer history (`NULLIF`).
    - Correct grouping of shots by `face_type` and `is_indoor`.
    - Euclidean distance math accuracy on `group_radius`.

### 6.2 Backend Unit & Integration Tests

- **Repository Tests (`internal/repository/reporting_test.go`):**
    - Verify all 6 analytical query functions invoke stored functions and unmarshal properly into Go
      structs.
- **Handler Tests (`internal/handler/analytics_test.go`):**
    - Verify `requireOwnership` returns `403 Forbidden` if archer ID in path does not match JWT
      token.
    - Verify query param validation for dates and enums.
    - Verify `200 OK` status and snake_case JSON serialization.
- Run `go test -race ./... -v` and `./scripts/linting.bash --go`.

### 6.3 Frontend Type Sync & Unit Tests

- Run `npm run generate:types` to regenerate `@/types/types.generated` from updated backend OpenAPI
  schema.
- **Vitest Unit Tests:**
    - `useAnalytics.spec.ts`: Test cached `archer_id` sync, URL query sync, and `AbortController`
      cancellation.
    - `AnalyticsFilterBar.spec.ts`: Test preset selections and filter emissions.
    - `ScoringTrendChart.spec.ts`: Test Average vs Total Score toggle and point click events.
    - `TargetDispersionView.spec.ts`: Test group tab switching and SVG element rendering.
    - `SessionDetailDrawer.spec.ts`: Test drawer open/close and keyboard ESC handling.
- Run `npm run type-check` (`vue-tsc --noEmit`) and `npm run lint`.

### 6.4 Responsive & UX Verification

- Start backend and Vite server.
- Verify smooth responsive layout across mobile (<768px), tablet, and desktop viewports.
- Verify dark mode aesthetics and chart contrast.

## 7. Delivery Milestones

1. **Milestone 1 (Database & Migrations):** Goose migration creating composite indexes and the 6
   analytical PL/pgSQL stored functions.
2. **Milestone 2 (Backend Services & Endpoints):** Models, `ReportingRepo`, `AnalyticsHandler`, Chi
   routes, auth ownership verification, and Go tests.
3. **Milestone 3 (OpenAPI & Type Generation):** Swagger annotations, OpenAPI generation, and
   frontend TypeScript type regeneration.
4. **Milestone 4 (Frontend State & Chart Components):** `Chart.js` setup, `useAnalytics.ts`
   composable with auth caching, and unit tests.
5. **Milestone 5 (Dashboard Assembly & Polishing):** `AnalyticsDashboard.vue`,
   `AnalyticsFilterBar.vue`, `ScoringTrendChart.vue`, `TargetDispersionView.vue`,
   `ScoreDistributionChart.vue`, and `SessionDetailDrawer.vue`.
6. **Milestone 6 (Verification & Review):** Automated test runs, linting checks, and visual
   responsive review.
