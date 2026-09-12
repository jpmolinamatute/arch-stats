# Task 002: Domain Model Structs in `internal/model/`

## Git Branch

`feature/002-domain-model-structs`

## Objective

Define domain model structs with JSON and validation tags for `bow` and `arrow`, refactor
`archer.go` to contain only personal information, and update `auth.go` domain models to
represent system-managed Google OAuth identity data.

## Dependencies

- Task 001 (Database migrations and schema)

## Acceptance Criteria

- [ ] New file `backend/internal/model/bow.go` defines:
    - [ ] `BowCreate`: `ArcherID`, `Name`, `Bowstyle`, `DrawWeight`
    - [ ] `BowRead`: `BowID`, `ArcherID`, `Name`, `Bowstyle`, `DrawWeight`, `IsDeleted`, `CreatedAt`
    - [ ] `BowSet`: nullable fields for updates (`Name`, `Bowstyle`, `DrawWeight`)
    - [ ] `BowFilter`: `BowID`, `ArcherID`, `Bowstyle`, `IsDeleted`
- [ ] New enum `ArrowStatus` in `backend/internal/model/enums.go`:
    - [ ] Values: `ArrowStatusInUse = "in_use"`, `ArrowStatusDamaged = "damaged"`, `ArrowStatusLost
          = "lost"`
- [ ] New file `backend/internal/model/arrow.go` defines:
    - [ ] `ArrowCreate`: `ArcherID`, `ArrowSet`, `ArrowNumber`, `Status` (optional, defaults to
          `in_use`), `Spine`, `Length`, `Weight`
    - [ ] `ArrowBatchCreate`: `ArcherID`, `ArrowSet`, `Count` (min 3), `Status` (optional, defaults
          to `in_use`), `Spine`, `Length`, `Weight`
    - [ ] `ArrowRead`: `ArrowID`, `ArcherID`, `ArrowSet`, `ArrowNumber`, `Status`, `Spine`, `Length`
          , `Weight`, `IsDeleted`, `CreatedAt`
    - [ ] `ArrowSet`: nullable fields for updates (`Status`, `Spine`, `Length`, `Weight`)
    - [ ] `ArrowFilter`: `ArrowID`, `ArcherID`, `ArrowSet`, `Status`, `IsDeleted`
- [ ] Modified `backend/internal/model/archer.go`:
    - [ ] Equipment fields removed (`Bowstyle`, `DrawWeight`)
    - [ ] Legacy fields removed (`ClubID`, `GoogleSubject`, `GooglePictureURL`, `LastLoginAt`,
          `CreatedAt`)
    - [ ] `ArcherCreate` requires: `Email`, `FirstName`, `LastName`, `DateOfBirth`, `Gender`
    - [ ] `ArcherRead` includes: `ArcherID`, `Email`, `FirstName`, `LastName`, `DateOfBirth`,
          `Gender`, `IsDeleted`
- [ ] Modified `backend/internal/model/auth.go`:
    - [ ] `AuthIdentityRead` represents OAuth record: `ArcherID`, `GoogleSubject`,
          `GooglePictureURL`, `LastLoginAt`, `CreatedAt`
    - [ ] `AuthStatus` enum values preserved: `authenticated`, `needs_registration`
    - [ ] `AuthNeedsRegistration` includes `GoogleEmail`, `GoogleSubject`, `GivenName`,
          `FamilyName`, `PictureURL`
- [ ] JSON tags match snake_case API specifications.
- [ ] Unit tests in `backend/internal/model/model_test.go` verify JSON marshaling and validation.
- [ ] `cd backend && go test ./internal/model/... -v` passes.
- [ ] `cd backend && go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/model/enums.go` |
| Create | `backend/internal/model/bow.go` |
| Create | `backend/internal/model/arrow.go` |
| Modify | `backend/internal/model/archer.go` |
| Modify | `backend/internal/model/auth.go` |
| Modify | `backend/internal/model/model_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [archer.go](../../../backend/internal/model/archer.go)
- [auth.go](../../../backend/internal/model/auth.go)
- [enums.go](../../../backend/internal/model/enums.go)

## Steps

- [ ] **Step 1: Write failing tests in `model_test.go`**

  Add tests for `BowRead`, `BowCreate`, `ArrowRead`, `ArrowBatchCreate`, updated `ArcherRead`,
  and `AuthIdentityRead` verifying:
    - Serialization to JSON produces exact expected snake_case keys.
    - Deserialization properly parses valid JSON payloads.
    - Validations flag negative draw weight or invalid enums.

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/model/... -v
  ```

- [ ] **Step 3: Implement `bow.go`**

  Create `backend/internal/model/bow.go`:

  ```go
  package model

  import (
      "time"

      "github.com/google/uuid"
  )

  type BowCreate struct {
      ArcherID   uuid.UUID `json:"archer_id"`
      Name       string    `json:"name" validate:"required,min=1,max=255"`
      Bowstyle   Bowstyle  `json:"bowstyle" validate:"required"`
      DrawWeight float64   `json:"draw_weight" validate:"required,gt=0"`
  }

  type BowRead struct {
      BowID      uuid.UUID `json:"bow_id"`
      ArcherID   uuid.UUID `json:"archer_id"`
      Name       string    `json:"name"`
      Bowstyle   Bowstyle  `json:"bowstyle"`
      DrawWeight float64   `json:"draw_weight"`
      IsDeleted  bool      `json:"is_deleted"`
      CreatedAt  time.Time `json:"created_at"`
  }

  type BowSet struct {
      Name       *string   `json:"name,omitempty"`
      Bowstyle   *Bowstyle `json:"bowstyle,omitempty"`
      DrawWeight *float64  `json:"draw_weight,omitempty"`
  }

  type BowFilter struct {
      BowID     *uuid.UUID `json:"bow_id,omitempty"`
      ArcherID  *uuid.UUID `json:"archer_id,omitempty"`
      Bowstyle  *Bowstyle  `json:"bowstyle,omitempty"`
      IsDeleted *bool      `json:"is_deleted,omitempty"`
  }
  ```

- [ ] **Step 4: Implement `arrow.go`**

  Create `backend/internal/model/arrow.go`:

  ```go
  package model

  import (
      "time"

      "github.com/google/uuid"
  )

  type ArrowCreate struct {
      ArcherID    uuid.UUID    `json:"archer_id"`
      ArrowSet    int16        `json:"arrow_set" validate:"required,gt=0"`
      ArrowNumber int16        `json:"arrow_number" validate:"required,gt=0"`
      Status      *ArrowStatus `json:"status,omitempty"`
      Spine       *float64     `json:"spine,omitempty"`
      Length      *float64     `json:"length,omitempty"`
      Weight      *float64     `json:"weight,omitempty"`
  }

  type ArrowBatchCreate struct {
      ArcherID uuid.UUID    `json:"archer_id"`
      ArrowSet int16        `json:"arrow_set" validate:"required,gt=0"`
      Count    int          `json:"count" validate:"required,gte=3"`
      Status   *ArrowStatus `json:"status,omitempty"`
      Spine    *float64     `json:"spine,omitempty"`
      Length   *float64     `json:"length,omitempty"`
      Weight   *float64     `json:"weight,omitempty"`
  }

  type ArrowRead struct {
      ArrowID     uuid.UUID   `json:"arrow_id"`
      ArcherID    uuid.UUID   `json:"archer_id"`
      ArrowSet    int16       `json:"arrow_set"`
      ArrowNumber int16       `json:"arrow_number"`
      Status      ArrowStatus `json:"status"`
      Spine       *float64    `json:"spine,omitempty"`
      Length      *float64    `json:"length,omitempty"`
      Weight      *float64    `json:"weight,omitempty"`
      IsDeleted   bool        `json:"is_deleted"`
      CreatedAt   time.Time   `json:"created_at"`
  }

  type ArrowSet struct {
      Status *ArrowStatus `json:"status,omitempty"`
      Spine  *float64     `json:"spine,omitempty"`
      Length *float64     `json:"length,omitempty"`
      Weight *float64     `json:"weight,omitempty"`
  }

  type ArrowFilter struct {
      ArrowID   *uuid.UUID   `json:"arrow_id,omitempty"`
      ArcherID  *uuid.UUID   `json:"archer_id,omitempty"`
      ArrowSet  *int16       `json:"arrow_set,omitempty"`
      Status    *ArrowStatus `json:"status,omitempty"`
      IsDeleted *bool        `json:"is_deleted,omitempty"`
  }
  ```

- [ ] **Step 5: Refactor `archer.go` and `auth.go`**

  Remove equipment and Google identity fields from `ArcherCreate`, `ArcherRead`, and `ArcherSet`.
  Include `IsDeleted bool json:"is_deleted"` in `ArcherRead`.
  Define `AuthIdentityRead` and updated `AuthNeedsRegistration` in `auth.go`.

- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/model/... -v
  golangci-lint run ./internal/model/...
  ```

- [ ] **Step 7: Commit changes**

  ```bash
  git add backend/internal/model
  git commit -m "feat(model): add bow and arrow models, refactor archer and auth models"
  ```

## Verification

- `cd backend && go test ./internal/model/... -v` exits with code 0.
- `cd backend && golangci-lint run ./internal/model/...` passes with zero errors.
- JSON field naming matches OpenAPI and database specifications.
