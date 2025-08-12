# Simplified PRD list

## Single-Page Application (SPA)

1. Check landing page
   * if sessionOpened.is_opened is true, display table shots. (use frontend/src/state/session.ts for this)
   * if sessionOpened.is_opened is true and the table shots is already displayed use websocket to pull live shots. 
   * if sessionOpened.is_opened is false, ask to open a session. (use frontend/src/state/session.ts for this)
2. Create Session page.
   * There must be an "open session" button at the top, if sessionOpened.is_opened is false.(use frontend/src/state/session.ts for this)
   * Opened session data (if any) will be display at the top, if sessionOpened.is_opened is true. There must be only one opened session at once. This should replace the "open session" button.
   * This session will pull all session data from the API server and list them in a table. Except any opened session.
   * Opened session data (if any) will be display at the top. There must be only one opened session at once.
3. Arrow registration page.
4. Target calibration page.
