---
trigger: glob
globs: backend/**/*.py
---

# New Code

## Coding

When suggesting new code follow these rules:

* Validate code against the latest documentation from the Context7 MCP.
* Functions and methods must be properly type annotated (avoid "Any" types at any cost).
* Code must pass linting. Use backend-linting skill.
* Code must pass type annotation checks. Use backend-type-annotation skill.
* Code must be properly formatted. Use backend-formatting skill.
* Always write tests for new code, write success and fail path tests and all tests must pass.
  Use backend-tests skill.

## Running Code or Python Scripts

We use uv for creating python virtual environments, managing python packages (aka dependencies)
and running scripts. In order to do this, please use backend-virtual-environment, backend-package-management
and backend-scripts skills.

## Feedback

* Explain what was done and why it was done
