---
trigger: glob
globs: backend/**/*.go
---

# New Go Backend Code

## Coding Standards

When creating or modifying Go backend code, follow these rules:

* Adhere to [Effective Go](https://go.dev/doc/effective_go) standards. Use the `backend-go-coding` skill.
* Code must pass linting and formatting without errors. Use the `backend-linting` skill (`cd backend && golangci-lint run ./...`).
* Always write unit tests for new code (covering both success and error paths). All tests must pass. Use the `backend-tests` skill (`cd backend && go test -race ./... -v`).
* Use `internal/apperror` sentinel errors and wrap errors using `%w` or `apperror.Wrap`.
* Do not introduce external dependencies without checking `backend-package-management`.

## Workflow Chain

Whenever writing Go code:
1. Implement code according to `backend-go-coding`.
2. Run `backend-linting` to format and check.
3. Run `backend-tests` to ensure tests pass.

## Feedback

* Explain what was done and why it was done in all agent responses.
