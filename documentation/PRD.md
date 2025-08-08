# **Arch-Stats Product Requirements Document (PRD)**

## 1. Problem Statement

Archers often lack objective, detailed feedback about their shooting performance. Arch-Stats solves this by combining hardware sensors with a real-time web interface to track every shot, visualize accuracy trends, and analyze arrow-specific performance.

## 2. Target Users

* **Primary:** Individual archers who want to measure improvement over time.
* **Secondary:** Coaches tracking a single athlete's performance.
* **Future:** Multi-archer teams and tournament organizers.

## 3. Goals & Success Metrics

| Goal                                            | Success Metric                                                     |
| ----------------------------------------------- | ------------------------------------------------------------------ |
| Provide real-time shot feedback                 | UI updates within 200ms of target sensor reading                   |
| Identify arrow-specific performance differences | At least 95% of registered arrows correctly identified in sessions |
| Track long-term performance trends              | View and compare data over time from database without errors       |

## 4. Current System Overview

### Architecture Summary

* **Backend** (FastAPI, asyncpg) - coordinates arrow registration, session tracking, and sensor data ingestion.

  * `src/arrow_reader/` → Handles arrow ID assignment
  * `src/bow_reader/` → Logs draw/release
  * `src/target_reader/` → Logs impact location & time
  * `src/server/` → REST + WebSocket API
* **Frontend** (Vue 3, TypeScript, Tailwind, Vite) - provides real-time visualizations and forms.
  * Key components: `WizardSession.vue`, `ShotsTable.vue`, `NewArrow.vue`
* **Database** (PostgreSQL 15+) - central store with `arrows`, `sessions`, `shots` `targets` tables.

## 5. Non-Functional Requirements

* **Performance:** End-to-end latency for shot → UI <200ms
* **Reliability:** Handle sensor disconnects gracefully
* **Security:** Only authenticated users can access or modify data
* **Deployment:** Must run on Raspberry Pi 5 (Linux) with Docker Compose

## 6. Future Enhancements (Global)

* Multi-archer profiles
* Advanced analytics (group size calculations, trend predictions)
* Cloud sync and remote coaching mode
