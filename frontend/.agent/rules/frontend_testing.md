---
trigger: glob
description: Use this rule whenever the user asks to run tests or verify frontend code integrity.
globs: frontend/**/*
---

# Frontend Testing Standards

Whenever you need to run tests, you must follow these steps:

1. **Change Directory:** Always change your working directory to `./frontend/`.
2. **Execute Command:** Run `npm run test`.
3. **Report:** Provide a summary of passing/failing tests to the user.
