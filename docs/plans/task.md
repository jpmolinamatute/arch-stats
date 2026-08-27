# Task List: Task 003 - Create `internal/config/` with envconfig

| Status | Task Description |
| :---: | :--- |
| `[x]` | Step 1: Create and checkout git branch `refactor/003-config-package` |
| `[x]` | Step 2: Add `envconfig` and `godotenv` dependencies in `backend/go.mod` |
| `[x]` | Step 3: Update `../.env` with missing fields and create `backend/.env` symlink |
| `[x]` | Step 4: Write failing unit tests in `backend/internal/config/config_test.go` |
| `[x]` | Step 5: Run `go test ./internal/config/...` to verify test failure |
| `[x]` | Step 6: Implement `backend/internal/config/config.go` |
| `[x]` | Step 7: Run `go test ./internal/config/... -v` to verify all tests pass |
| `[x]` | Step 8: Run `go vet ./...` and `go build ./...` across backend |
| `[x]` | Step 9: Remove `backend/internal/config/.gitkeep` |
| `[x]` | Step 10: Commit changes to `refactor/003-config-package` |
