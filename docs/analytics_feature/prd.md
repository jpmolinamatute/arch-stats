# Product Requirements Document (PRD) & Technical Specification

## Historical Analytics & Visual Charting Feature

- **Target Route:** `/app/analytics`
- **Feature Owner:** Analytics & Frontend / Backend Core
- **Status:** Approved / Design Complete
- **Date:** 2026-09-08
- **Storage Location:** `docs/analytics_feature/prd.md`

---

## 1. Executive Summary & Goals

### 1.1 Problem Statement

In archery, tracking performance over time is essential for equipment tuning, form consistency, and tournament preparation. Currently, the platform collects real-time shot coordinates and scores during live shooting sessions, but lacks a centralized, historical analytics hub where archers can visualize long-term progression, analyze target dispersion, diagnose arrow equipment, and measure stamina across sessions, distances, and bowstyles.

### 1.2 Core Value Proposition

Provide archers with an intuitive, visually stunning, and responsive **Historical Analytics Dashboard** (`/app/analytics`) that answers four critical questions:

1. **"Am I improving?"** (Scoring progression trendline, 5-session moving averages, and ring hit distribution).
2. **"Where do my arrows cluster, and is my sight centered?"** (Target face dispersion, Center of Mass drift, group radius, and horizontal vs. vertical form spread).
3. **"Are specific arrows hurting my score?"** (Arrow shaft diagnostics to isolate damaged nocks, vanes, or flyers among shafts #1–#12).
4. **"When does fatigue set in?"** (End-by-end endurance curve tracking score drop-off and cluster degradation from End 1 to End 10).

### 1.3 Key Architectural Decisions

- **PostgreSQL-First Computation:** All complex data manipulation, math (Center of Mass, group spread radius, horizontal/vertical standard deviation, shaft aggregation, moving averages, and end fatigue) is encapsulated in PostgreSQL 17 stored functions via Goose migrations. The Go Chi backend server remains thin, executing simple `SELECT * FROM get_...()` queries.
- **Normalized Multi-Dimensional Filtering:** Shots are evaluated strictly within comparable disciplines (Distance, Target Face, Bowstyle, Indoor/Outdoor, Location, Date Range) to eliminate artificial variance.
- **Granular Micro-Endpoints & Tab-Aware Lazy Loading:** Independent REST endpoints (`/filters`, `/kpis`, `/progression`, `/distribution`, `/dispersion`, `/arrows`, `/fatigue`, `/session/{id}`) allow the frontend to fetch summary KPIs globally while lazily loading heavy visualization datasets only when their corresponding tab is active.
- **Zero-Latency Auth Caching:** `archer_id` is cached synchronously in `localStorage` upon login, allowing the dashboard to bootstrap and fire initial analytics queries immediately on mount without waiting for `/auth/me`.
- **Hybrid Charting Engine:** Chart.js via `vue-chartjs` powers time-series trends, histograms, arrow diagnostics, and fatigue curves; the SVG engine powers target face dispersion and sight drift overlays.

---

## 2. Dimensional Data Taxonomy (Schema Alignment)

Based on the entities in `backend/migrations` (`archer`, `session`, `target`, `slot`, `arrow`, `shot`), the analytics architecture strictly separates **Filters (Dimensions)** from **Metrics (Measures)**:

```text
archer (who) ──▶ session (where/when) ──▶ target (distance) ──▶ slot (face/bow) ──▶ shot (score/coords/arrow)
```

### 2.1 Filters & Dimensions (Context & Slicing)

| Dimension | Migration Source | Possible Values | Analytical Purpose |
| :--- | :--- | :--- | :--- |
| **Date Range** | `session.created_at` | Presets (`7d`, `30d`, `90d`, `season`, `all`) or Custom | Time-bounding analysis periods. |
| **Distance** | `target.distance` | Integer (e.g. 18m, 30m, 50m, 70m) | Normalizes target distance (prevents mixing 18m and 70m). |
| **Target Face** | `slot.face_type` | `wa_40cm_full`, `wa_40cm_triple_vertical`, `wa_122cm_full`, etc. | Dictates ring geometry and scoring dimensions. |
| **Bowstyle** | `slot.bowstyle` | `recurve`, `compound`, `barebow`, `longbow` | Isolates equipment style. |
| **Environment** | `session.is_indoor` | `true` (Indoor), `false` (Outdoor) | Isolates wind/weather impact from indoor conditions. |
| **Location / Venue**| `session.session_location`| String (e.g. Club, Range, Home, Competition) | Evaluates performance across different venues. |
| **Arrow Number** | `arrow.arrow_number` | Integer (`#1` through `#12`) | Shaft-level grouping for equipment diagnostics. |
| **End Index** | Derived: order within `slot` | Integer (`End 1` through `End 10`) | Within-session chronological progression for fatigue. |

### 2.2 Metrics & Measures (Aggregated Signals)

| Metric | Calculation | Archery Meaning |
| :--- | :--- | :--- |
| **Arrow Average** | `AVG(shot.score)` (0.00 – 10.00) | Overall scoring average normalized across session lengths. |
| **Total Volume** | `COUNT(shot.shot_id)` | Total arrows logged in the filtered scope. |
| **X-Count & X-Rate**| `COUNT(is_x)` and `(COUNT(is_x) / COUNT(*)) * 100` | Elite accuracy indicator. |
| **10-Rate** | `(COUNT(score = 10) / COUNT(*)) * 100` | Gold-ring consistency. |
| **Personal Best** | `MAX(session_total_score)` | Best single session performance for the discipline. |
| **Center of Mass** | $(\overline{x}, \overline{y}) = (\text{AVG}(x), \text{AVG}(y))$ | **Sight / Wind Bias:** Grouping offset from center (sight pin adjustment). |
| **Group Radius** | $\max \sqrt{(x - \overline{x})^2 + (y - \overline{y})^2}$ | **Precision (Cluster Tightness):** Form consistency irrespective of sight zero. |
| **Horizontal Spread**| $\sigma_x = \text{STDDEV}(x)$ | **Form Flaw:** Bow hand torque, grip pressure, or crosswinds. |
| **Vertical Spread** | $\sigma_y = \text{STDDEV}(y)$ | **Form Flaw:** Draw length inconsistency, anchor variation, shoulder collapse. |
| **Shaft Variance** | $\overline{\text{score}}_{\text{arrow}} - \overline{\text{score}}_{\text{quiver}}$ | Identifies rogue "flyer" arrows. |
| **Fatigue Delta** | $\text{AVG}(\text{score})_{\text{Ends 1-3}} - \text{AVG}(\text{score})_{\text{Ends 7-10}}$ | Stamina degradation indicator. |

---

## 3. User Experience & Dashboard Architecture (Approach 1: Unified Tabbed Analytics Hub)

### 3.1 Information Architecture & Layout

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                    AppHeader.vue (Navigation Link)                       │
└──────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    AnalyticsDashboard.vue (/app/analytics)               │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 1. AnalyticsFilterBar.vue (Sticky Top Bar with URL Sync)           │  │
│  │    [Date Range ▾] [Distance ▾] [Face ▾] [Bowstyle ▾] [In/Out ▾] ... │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 2. AnalyticsKPIDeck.vue (4 Hero Cards with Skeleton Loaders)       │  │
│  │    [ Avg Arrow Score ] [ Total Arrows ] [ Group Spread ] [ PB ]    │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 3. AnalyticsTabBar.vue (Navigation Pills)                          │  │
│  │    [ 📈 Progression & Rings ] [ 🎯 Target & Dispersion ]           │  │
│  │    [ 🏹 Arrow Tuning ]        [ ⏱️ Fatigue Curve ]                 │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ Dynamic Active Tab View:                                           │  │
│  │                                                                    │  │
│  │ • Tab 1 (ProgressionView.vue):                                     │  │
│  │     - ScoringTrendChart.vue (Timeline Spline + 5-Session MA)       │  │
│  │     - ScoreDistributionChart.vue (WA Ring Color Histogram X-M)     │  │
│  │                                                                    │  │
│  │ • Tab 2 (TargetDispersionView.vue):                                │  │
│  │     - SVG Target Face (Plotted shots + Center of Mass Crosshair)   │  │
│  │     - Group Radius Circle + Diagnostic Badges (σx vs. σy spread)   │  │
│  │                                                                    │  │
│  │ • Tab 3 (ArrowDiagnosticsView.vue - "Flyer Finder"):               │  │
│  │     - ArrowScoreBarChart.vue (Score per shaft #1–#12 + Alert Tag)  │  │
│  │     - Target Quadrant Shaft Overlay                                │  │
│  │                                                                    │  │
│  │ • Tab 4 (FatigueAnalysisView.vue):                                 │  │
│  │     - EndProgressionChart.vue (Score & Group Spread by End 1–10)   │  │
│  │     - Endurance Drop-Off Badge                                     │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 4. SessionDetailDrawer.vue (Slide-Out Side Panel / Bottom Sheet)   │  │
│  │    - Scorecard table by end + single-session mini target SVG       │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Component Specifications

#### 3.2.1 `AnalyticsFilterBar.vue` (Primary Sticky Filter Bar)
- **Positioning:** Sticky below `AppHeader.vue`.
- **Controls:**
  1. **Date Range Dropdown:** Presets ("Last 7 Days", "Last 30 Days", **"Last 90 Days" (Default)**, "This Season", "All Time", "Custom Range").
  2. **Distance Dropdown:** Options: "All Distances" (Default) plus dynamically loaded distinct distances shot by archer (e.g., 18m, 50m, 70m).
  3. **Target Face Dropdown:** Options: "All Faces" (Default) plus distinct face types (e.g. `wa_40cm_full`, `wa_122cm_full`).
  4. **Bowstyle Dropdown:** Options: "All Bowstyles" (Default), "Recurve", "Compound", "Barebow", "Longbow".
  5. **Environment Dropdown:** Options: "All Environments" (Default), "Indoor Only", "Outdoor Only".
  6. **Location Dropdown:** Options: "All Locations" (Default) plus distinct venues shot.
- **Two-Way URL Synchronization:** All active filters and the active tab sync to URL query parameters (`?tab=progression&range=90d&distance=18&face=wa_40cm_full&bowstyle=recurve`).

#### 3.2.2 `AnalyticsKPIDeck.vue` (Hero Cards)
- **Card 1: Arrow Average:** Mean score per arrow (e.g. `9.28 / 10`).
- **Card 2: Total Arrows Shot:** Arrow volume logged in period (e.g. `1,420`).
- **Card 3: Group Spread:** Average group spread diameter (e.g. `48.2 mm`).
- **Card 4: Personal Best:** Highest single-session score in period (e.g. `294 / 300`).
- **States:** Shimmer skeleton animation while queries resolve; `--` fallback if no data.

#### 3.2.3 `ProgressionView.vue` (Tab 1)
- **`ScoringTrendChart.vue`:**
  - Chart.js spline line chart connecting chronologically ordered sessions.
  - Header segmented toggle: `Average Arrow Score (0.00 – 10.00)` (default) vs. `Total Session Score`.
  - 5-session windowed moving average dashed line.
  - Tooltip: Date, Location, Distance, Face, Total Arrows, Arrow Average, Session Total.
  - Clicking any session node emits `select-session(sessionId)` to trigger the detail drawer.
- **`ScoreDistributionChart.vue`:**
  - Horizontal bar chart colored to World Archery ring colors (Gold: X/10/9, Red: 8/7, Blue: 6/5, Black: 4/3, White: 2/1, Gray: M/0).
  - Displays exact hit counts and percentages per ring.

#### 3.2.4 `TargetDispersionView.vue` (Tab 2)
- **SVG Target Face Renderer:** Dynamically loads the target face SVG corresponding to active `face_type` filter.
- **Shot Impacts:** Plots all `(x, y)` shot coordinates as semi-transparent dots with opacity density blending.
- **Center of Mass Crosshair:** Plotted at $(\overline{x}, \overline{y})$ indicating sight or wind drift.
- **Group Radius Circle:** Circle centered on $(\overline{x}, \overline{y})$ with radius $R_{max}$.
- **Form Diagnostics Badge Deck:**
  - `Group Spread (mm)`: Overall cluster diameter.
  * `Vertical Spread (σy)`: Anchor, draw length, or bow shoulder stability.
  * `Horizontal Spread (σx)`: Grip pressure, bow torque, or wind resistance.

#### 3.2.5 `ArrowDiagnosticsView.vue` (Tab 3: "Flyer Finder")
- **`ArrowScoreBarChart.vue`:**
  - Bar chart with X-axis showing shaft numbers (`#1` through `#12`) and Y-axis showing average score per shaft.
  - Dotted horizontal benchmark line indicating the quiver-wide average score.
  - Shafts performing $\ge 1.5$ standard deviations below quiver average are highlighted with an amber alert badge (*"Potential Flyer: Arrow #4 averages 1.3 pts lower"*).
- **Target Face Shaft Overlay:** Toggle to color target impact points by arrow number to reveal quadrant-specific drift.

#### 3.2.6 `FatigueAnalysisView.vue` (Tab 4: Endurance Curve)
- **`EndProgressionChart.vue`:**
  - Dual-axis line chart: X-axis is End Number (`End 1` to `End 10`).
  - Primary Y-axis: Average score per end.
  - Secondary Y-axis: Group spread radius (mm) per end.
- **Endurance Drop-off Badge:** Displays the scoring delta between the first third and final third of sessions (e.g. `-0.38 pts/arrow in 2nd half`).

#### 3.2.7 `SessionDetailDrawer.vue` (Drill-Down Drawer)
- Slide-over right drawer (desktop) / bottom sheet (mobile) with backdrop dismissal.
- Header: Session Date, Venue, Distance, Face Type, Bowstyle, Total Arrows, Final Score.
- Mini SVG target face showing individual session shots.
- End-by-end scorecard table (e.g., End 1: 10, 10, 9 = 29).

---

## 4. System Architecture & Database Layer

### 4.1 Database Migrations (Goose)

A new migration `backend/migrations/007_2026-09-08_analytics_functions.sql` defines:

#### 4.1.1 Composite Indexes

```sql
CREATE INDEX IF NOT EXISTS idx_session_owner_created ON session(owner_archer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_target_session_dist ON target(session_id, distance);
CREATE INDEX IF NOT EXISTS idx_slot_archer_target ON slot(archer_id, target_id, face_type, bowstyle);
CREATE INDEX IF NOT EXISTS idx_shot_slot_arrow_score ON shot(slot_id, arrow_id, score, is_x, created_at);
```

#### 4.1.2 PostgreSQL Stored Functions

1. **`get_archer_analytics_filters(p_archer_id UUID)`**
   - Returns distinct filter values: `distances INT[]`, `face_types TEXT[]`, `bowstyles TEXT[]`, `locations TEXT[]`, `has_indoor BOOLEAN`, `has_outdoor BOOLEAN`, `min_date TIMESTAMPTZ`, `max_date TIMESTAMPTZ`.
2. **`get_archer_kpis(p_archer_id UUID, p_from TIMESTAMPTZ, p_to TIMESTAMPTZ, p_distance INT, p_face_type FACE_TYPE, p_bowstyle BOWSTYLE_TYPE, p_is_indoor BOOLEAN, p_location TEXT)`**
   - Computes: `total_sessions`, `total_shots`, `average_score`, `personal_best`, `x_count`, `x_rate`, `overall_group_spread_mm`.
3. **`get_archer_progression(p_archer_id UUID, ...filters)`**
   - Aggregates by session ordered by `session.created_at ASC`.
   - Computes windowed 5-session moving average: `AVG(avg_score) OVER (ORDER BY created_at ROWS BETWEEN 4 PRECEDING AND CURRENT ROW)`.
   - Returns rows: `session_id`, `date`, `distance`, `face_type`, `total_shots`, `total_score`, `average_score`, `x_count`, `moving_avg_5`.
4. **`get_archer_distribution(p_archer_id UUID, ...filters)`**
   - Groups by `shot.score` and `shot.is_x`.
   - Returns counts and percentages for rings `"X"`, `"10"`, `"9"` through `"0"`.
5. **`get_archer_dispersion(p_archer_id UUID, ...filters)`**
   - Filters shots where `x IS NOT NULL AND y IS NOT NULL`.
   - Computes:
     - `center_x = AVG(shot.x)`, `center_y = AVG(shot.y)`
     - `group_radius = COALESCE(MAX(SQRT(POWER(shot.x - cx, 2) + POWER(shot.y - cy, 2))), 0.0)`
     - `stddev_x = COALESCE(STDDEV(shot.x), 0.0)`, `stddev_y = COALESCE(STDDEV(shot.y), 0.0)`
     - `shots = json_agg(json_build_object('x', shot.x, 'y', shot.y, 'score', shot.score, 'is_x', shot.is_x, 'arrow_number', arrow.arrow_number))`
6. **`get_archer_arrow_diagnostics(p_archer_id UUID, ...filters)`**
   - Filters shots where `arrow_id IS NOT NULL`.
   - Groups by `arrow.arrow_number`.
   - Returns rows: `arrow_number`, `shot_count`, `average_score`, `center_x`, `center_y`, `group_radius`, `is_flyer`.
7. **`get_archer_fatigue(p_archer_id UUID, ...filters)`**
   - Derives `end_index` using integer division over ordered shots within slot.
   - Groups by `end_index`.
   - Returns rows: `end_index`, `shot_count`, `average_score`, `group_radius`.
8. **`get_archer_session_detail(p_archer_id UUID, p_session_id UUID)`**
   - Returns full denormalized session object: metadata, target distance, slot details, scorecard by end, and shot coordinate array.

---

## 5. Backend API Specification

All endpoints are mounted on Chi router under `/api/v1/analytics/archer/{archer_id}` and protected by `requireOwnership(w, r)`.

### 5.1 Standardized Filter Query Parameters

Endpoints accept standardized query parameters:
- `from` (ISO 8601 string, optional)
- `to` (ISO 8601 string, optional)
- `distance` (integer, optional)
- `face_type` (string enum, optional)
- `bowstyle` (string enum: `recurve`, `compound`, `barebow`, `longbow`, optional)
- `is_indoor` (boolean, optional)
- `location` (string, optional)

### 5.2 Micro-Endpoints

| HTTP Method & Path | Description | Response Payload Summary |
| :--- | :--- | :--- |
| `GET /api/v1/analytics/archer/{id}/filters` | Filter dropdown options | `{ distances: [18, 70], face_types: [...], bowstyles: [...], locations: [...], has_indoor, has_outdoor, min_date, max_date }` |
| `GET /api/v1/analytics/archer/{id}/kpis` | Hero summary cards | `{ total_sessions: 24, total_shots: 1420, average_score: 9.28, personal_best: 294, group_spread_mm: 48.2, x_rate: 46.8 }` |
| `GET /api/v1/analytics/archer/{id}/progression`| Tab 1: Chronological timeline | `[ { session_id, date, distance, face_type, total_shots, total_score, average_score, x_count, moving_avg_5 } ]` |
| `GET /api/v1/analytics/archer/{id}/distribution`| Tab 1: Ring accuracy histogram | `{ total_shots, counts: { "X": 665, "10": 280 ... }, percentages: { "X": 46.8 ... } }` |
| `GET /api/v1/analytics/archer/{id}/dispersion` | Tab 2: Group dispersion & drift | `{ center_x, center_y, group_radius, stddev_x, stddev_y, shots: [ { x, y, score, is_x, arrow_number } ] }` |
| `GET /api/v1/analytics/archer/{id}/arrows` | Tab 3: Shaft flyer diagnostics | `[ { arrow_number: 1, shot_count: 120, average_score: 9.42, center_x, center_y, group_radius, is_flyer: false }, ... ]` |
| `GET /api/v1/analytics/archer/{id}/fatigue` | Tab 4: End endurance curve | `[ { end_index: 1, shot_count: 240, average_score: 9.45, group_radius: 42.1 }, ... ]` |
| `GET /api/v1/analytics/archer/{id}/session/{sid}`| Drill-down session drawer | `{ session_id, date, distance, face_type, total_score, average_score, ends: [...], shots: [...] }` |

---

## 6. Frontend Architecture & State Management

### 6.1 Synchronous Auth Bootstrap

- In `frontend/src/composables/useAuth.ts`:
  - `archer_id` is cached synchronously in `localStorage.setItem('arch-stats:archer_id', archerId)`.
  - On logout, `localStorage.removeItem('arch-stats:archer_id')`.
- `useAnalytics.ts` reads `archer_id` synchronously on mount, firing `/filters` and `/kpis` with zero roundtrip delay.

### 6.2 Composable `useAnalytics.ts`

- **Reactive State:**
  - `activeTab`: `'progression' | 'dispersion' | 'arrows' | 'fatigue'`
  - `filters`: `{ range: '90d', startDate: null, endDate: null, distance: null, faceType: null, bowstyle: null, isIndoor: null, location: null }`
  - `filterOptions`: `{ distances: [], faceTypes: [], bowstyles: [], locations: [], hasIndoor: true, hasOutdoor: true }`
  - `kpis`: `Ref<ArcherKPIReport | null>`
  - `progression`: `Ref<SessionTrendPoint[]>`
  - `distribution`: `Ref<ScoreDistribution | null>`
  - `dispersion`: `Ref<TargetDispersionReport | null>`
  - `arrowDiagnostics`: `Ref<ArrowDiagnosticReport[]>`
  - `fatigue`: `Ref<FatiguePoint[]>`
  - `selectedSession`: `Ref<SessionDetailReport | null>`
  - Granular loading flags: `loadingFilters`, `kpisLoading`, `tabLoading`, `sessionDetailLoading`.
- **Tab-Aware Lazy Fetching:**
  - Changing filters immediately triggers `/kpis` and the active tab's endpoint(s).
  - Switching tabs checks if the target tab's dataset is loaded for the current filter hash; if not, fetches it on demand.
  - All calls cancel previous in-flight requests via `AbortController`.

---

## 7. Edge Cases & Error Handling

- **New Archer (Zero Sessions):** Renders onboarding card: *"No shooting sessions logged yet. Record your first round to unlock historical analytics"* with CTA routing to `/app`.
- **No Filter Matches:** Renders an inline banner: *"No sessions match this filter combination"* with a *"Reset Filters"* button.
- **Unassigned / Unnumbered Arrows:** Tab 3 displays an educational empty state explaining how numbering shafts (#1–#12) unlocks flyer detection and shaft tuning.
- **Single-Shot / Zero Variance:** Returns `group_radius = 0.0` and `stddev = 0.0` using SQL `COALESCE` to prevent division-by-zero or `NaN`.
- **End Size Normalization:** Dynamic grouping on `slot.shot_per_round` ensures 3-arrow and 6-arrow ends aggregate cleanly into relative end progression indices.

---

## 8. Testing & Verification Plan

### 8.1 Database Verification
- Execute Goose migration `007_2026-09-08_analytics_functions.sql` up and down.
- Verify division-by-zero protection (`NULLIF`) and Euclidean distance math on test datasets.

### 8.2 Backend Unit & Integration Tests
- `internal/repository/reporting_test.go`: Verify all 8 stored functions unmarshal accurately into Go domain structs.
- `internal/handler/analytics_test.go`:
  - Verify `requireOwnership` returns `403 Forbidden` on mismatched JWTs.
  - Verify parameter validation for date ranges and enums.
  - Verify JSON contracts match OpenAPI schema.
- Run `go test -race ./... -v` and `./scripts/linting.bash --go`.

### 8.3 Frontend Type Sync & Vitest Tests
- Run `npm run generate:types` to regenerate TypeScript types from the updated OpenAPI spec.
- Vitest unit tests:
  - `useAnalytics.spec.ts`: Test URL query synchronization, lazy tab fetching, and `AbortController` cancellation.
  - Component tests for `AnalyticsFilterBar.vue`, `AnalyticsKPIDeck.vue`, `ScoringTrendChart.vue`, `TargetDispersionView.vue`, `ArrowScoreBarChart.vue`, and `EndProgressionChart.vue`.
- Run `npm run type-check` (`vue-tsc --noEmit`) and `npm run lint`.

---

## 9. Delivery Milestones

1. **Milestone 1 (Database & Migrations):** Goose migration `007_2026-09-08_analytics_functions.sql` defining composite indexes and the 8 PL/pgSQL stored functions.
2. **Milestone 2 (Backend Core):** Models, `ReportingRepo`, `AnalyticsHandler`, Chi router mounting, ownership checks, and Go integration tests.
3. **Milestone 3 (OpenAPI & Types):** Swaggo OpenAPI annotations, schema regeneration, and frontend TypeScript type sync via `npm run generate:types`.
4. **Milestone 4 (Frontend State & Chart Components):** `Chart.js` setup, `useAnalytics.ts` composable with auth caching and tab-aware lazy loading, and unit tests.
5. **Milestone 5 (Dashboard Assembly & Tabs):**
   - `AnalyticsDashboard.vue` & `AnalyticsFilterBar.vue`
   - Tab 1: `ProgressionView.vue` (`ScoringTrendChart.vue` + `ScoreDistributionChart.vue`)
   - Tab 2: `TargetDispersionView.vue` (SVG target + Center of Mass + group radius)
   - Tab 3: `ArrowDiagnosticsView.vue` (Flyer finder)
   - Tab 4: `FatigueAnalysisView.vue` (Endurance curve)
   - `SessionDetailDrawer.vue` (Scorecard drill-down)
6. **Milestone 6 (Verification & Polish):** Automated test runs, linting checks, mobile responsive validation, and user review.
