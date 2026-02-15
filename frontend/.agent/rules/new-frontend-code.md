---
trigger: glob
globs: frontend/**/*
---

# New Code

## Coding

When suggesting new code follow this rules:

* Validate code against the latest documentation from the Context7 MCP.
* Functions and methods must be properly type annotated (avoid "any" or "unknown" types at any cost).
   Use types from frontend/src/types/types.generated.ts.
* Code must pass linting and be properly formatted. Use frontend-linting skill.
* Always write tests for new code, write success and fail path tests and  All tests must pass.
  Use frontend-tests skill.
* Always generate types in order to keep backend and frontend in sync.
  Use frontend-type-generation skill.

## Feedback

* Explain what was done and why was done
