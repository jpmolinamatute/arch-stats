---
name: backend-package-management
description: How to manage Go module dependencies using go get, go mod tidy, and go mod download
---

# Go Package & Dependency Management

We use Go modules (`go.mod` and `go.sum` in `backend/`) to manage backend dependencies.

All commands should be executed from the `backend/` directory:

```bash
cd backend
```

## Add or Update Dependencies

```bash
# Add or upgrade a dependency
go get github.com/example/pkg@latest

# Add a specific version
go get github.com/example/pkg@v1.2.3
```

## Clean and Prune Dependencies

Always run `go mod tidy` after adding or removing package imports to remove unused modules and
update checksums in `go.sum`:

```bash
go mod tidy
```

## Download Cached Dependencies

```bash
go mod download
```

## Verify Dependencies

```bash
go mod verify
```
