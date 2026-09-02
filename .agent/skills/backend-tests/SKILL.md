---
name: backend-tests
description: How to run Go backend unit and integration tests using go test
---

# Go Backend Tests

We use Go's standard testing toolchain (`go test`) for unit and integration testing in `backend/`.

## Running Tests

All commands are executed from the `backend/` directory:

```bash
cd backend
```

### 1. Run All Tests

```bash
go test ./... -v
```

### 2. Run with Race Detection (Recommended before commit)

```bash
go test -race ./... -v
```

### 3. Run Tests for a Specific Package

```bash
go test ./internal/config/... -v
go test ./internal/apperror/... -v
go test ./internal/repository/... -v
```

### 4. Run a Specific Test Function

```bash
go test ./internal/config/... -run TestNewLogger_DevMode -v
```

### 5. Bypass Test Cache

```bash
go test ./... -count=1 -v
```

### 6. Full Suite via Root Script

From the project root:

```bash
./scripts/linting.bash --go
```
