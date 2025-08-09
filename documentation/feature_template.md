# Feature Requirements Template

## Feature Name

**Example:** Session Tracking

**Description:**
Track the start, progression, and end of an archery practice session, logging all shots and metadata.

**User Story:**

> As an archer, I want to track my shooting sessions so that I can measure my performance across different days and conditions.

**Acceptance Criteria:**

1. User can open a session from WebUI.
2. Backend creates a new session in `sessions` table with `is_opened=true`.
3. Shots are logged automatically via bow and target sensors.
4. Session can be closed manually by user; `is_opened=false`, `end_time` set.

**Current Code References:**

* Backend: `src/server/routes/sessions.py`, `src/bow_reader/`, `src/target_reader/`
* Frontend: `components/WizardSession.vue`, `state/session.ts`
* Database: `sessions` and `shots` tables

**Dependencies:**

* Bow and target sensors online
* Frontend WebSocket connection active
* Arrow reader sensor connected and functional
* Backend WebSocket/REST endpoints online
  
**Future Enhancements:**

* Support for tournament scoring formats
* Integration with weather APIs for environmental data
* Bulk import/export of arrow data
* Photo attachment for each arrow
