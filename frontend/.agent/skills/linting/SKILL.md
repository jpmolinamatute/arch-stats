---
name: frontend-linting
trigger: glob
description: Apply this rule when saving files, fixing style issues, or when the user asks to "format" or "lint" the frontend.
globs: frontend/**/*
---

# Frontend Style Standards

To maintain code quality and consistency:

1. **Change Directory:** Navigate to the `./frontend/` directory.
2. **Execute Command:** Run `npm run lint` to both check for errors and apply automatic formatting.
3. **Constraint:** Do not bypass linting errors; if `npm run lint` fails, fix the
underlying code issues before finishing the task.
