# Login / Registration flow

We must use Google One Tap sign-in/sign-up.

When Page first loads

- check /api/v0/auth/ping.
- if returns false, the app does nothing (we wait for user to click on the login button)
- if returns true, the app auto login the user (auto_select: true to silently re-authenticate users who are still logged into Google.)
