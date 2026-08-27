# Task 001: Initialize Go Module and Project Directory Structure

## Git Branch

`refactor/001-go-module-and-project-scaffold`

## Objective

Create the new `backend/` directory with the Go module initialized (`go.mod`) and the full
directory skeleton defined in the high-level plan. Include a minimal `main.go` that compiles
and prints a startup message, proving the scaffold is valid.

## Dependencies

- Task 000 (backend renamed to backend-old)

## Acceptance Criteria

- [x] `backend/go.mod` exists with module path `github.com/jpmolinamatute/arch-stats/backend`
  and Go version `1.27.0`.
- [x] The following directory structure exists (with `.gitkeep` files to preserve empty dirs):

  ```txt
  backend/
  ├── go.mod
  ├── cmd/
  │   └── arch-stats/
  │       └── main.go
  └── internal/
      ├── apperror/
      ├── auth/
      ├── config/
      ├── handler/
      ├── middleware/
      ├── model/
      ├── repository/
      ├── service/
      └── websocket/
  ```

- [x] `go build ./cmd/arch-stats` succeeds with zero errors.
- [x] Running the binary prints: `arch-stats server starting...` and exits cleanly.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/go.mod` |
| Create | `backend/cmd/arch-stats/main.go` |
| Create | `backend/internal/apperror/.gitkeep` |
| Create | `backend/internal/auth/.gitkeep` |
| Create | `backend/internal/config/.gitkeep` |
| Create | `backend/internal/handler/.gitkeep` |
| Create | `backend/internal/middleware/.gitkeep` |
| Create | `backend/internal/model/.gitkeep` |
| Create | `backend/internal/repository/.gitkeep` |
| Create | `backend/internal/service/.gitkeep` |
| Create | `backend/internal/websocket/.gitkeep` |

## Steps

- [x] **Step 1: Initialize the Go module**

  ```bash
  mkdir -p backend
  cd backend
  go mod init github.com/jpmolinamatute/arch-stats/backend
  ```

  Edit `go.mod` to ensure the Go version line reads `go 1.27.0`.

- [x] **Step 2: Create the directory skeleton**

  ```bash
  mkdir -p cmd/arch-stats
  mkdir -p internal/{apperror,auth,config,handler,middleware,model,repository,service,websocket}
  ```

  Add `.gitkeep` files in each empty `internal/` subdirectory:

  ```bash
  for dir in internal/*/; do touch "$dir/.gitkeep"; done
  ```

- [x] **Step 3: Write the minimal `main.go`**

  Create `backend/cmd/arch-stats/main.go`:

  ```go
  package main

  import "fmt"

  func main() {
      fmt.Println("arch-stats server starting...")
  }
  ```

- [x] **Step 4: Verify it compiles**

  ```bash
  cd backend
  go build ./cmd/arch-stats
  ```

  Expected: exits 0, produces `arch-stats` binary in `backend/`.

- [x] **Step 5: Run the binary**

  ```bash
  ./arch-stats
  ```

  Expected output: `arch-stats server starting...`

- [x] **Step 6: Run go vet**

  ```bash
  go vet ./...
  ```

  Expected: no issues reported.

- [x] **Step 7: Clean up and commit**

  ```bash
  rm -f arch-stats  # remove compiled binary
  echo "arch-stats" >> .gitignore  # ignore compiled binary
  git add -A
  git commit -m "feat: initialize Go module and project directory scaffold"
  ```

## Verification

- `cat backend/go.mod` shows the module path and Go 1.27.0.
- `cd backend && go build ./cmd/arch-stats && ./arch-stats` prints the startup message.
- `cd backend && go vet ./...` reports clean.
- All `internal/*/` subdirectories exist.
