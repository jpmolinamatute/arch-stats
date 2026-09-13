# Domain Model Structs in `internal/model/` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement domain model structs with JSON and validation tags for `bow` and `arrow`, refactor `archer.go` to store strictly personal profile data, and update `auth.go` to support Google OAuth identity persistence.

**Architecture:** Add isolated Go structs and enums in `backend/internal/model/` (`enums.go`, `bow.go`, `arrow.go`, `archer.go`, `auth.go`) with canonical `snake_case` JSON tags and `go-playground/validator` tags. Verify full serialization, deserialization, and structural constraints with comprehensive unit tests in `model_test.go`.

**Tech Stack:** Go 1.27+, `github.com/google/uuid`, standard library `encoding/json` and `time`.

**Spec:** [docs/onboarding_refactor/tasks/002_domain_model_structs.md](../onboarding_refactor/tasks/002_domain_model_structs.md)

## Global Constraints

- Domain model package is strictly `backend/internal/model`.
- Go naming conventions: PascalCase for structs and fields, snake_case for JSON tags.
- Use `github.com/google/uuid` for UUID fields.
- Use pointers for optional/nullable fields (`*float64`, `*string`, `*ArrowStatus`, etc.) with `omitempty` JSON tag.
- Verification commands:
  - `cd backend && go test ./internal/model/... -v` must exit 0.
  - `cd backend && go vet ./internal/model/...` must report no issues.
  - `cd backend && golangci-lint run ./internal/model/...` must report 0 issues.
- All steps must follow strict Test-Driven Development (TDD): write failing tests first, verify failure, implement minimal code, verify pass, commit.
- Mark all acceptance criteria and step checkboxes in `docs/onboarding_refactor/tasks/002_domain_model_structs.md` as done (`[x]`) upon completion.
- Maintain a table-only live tracker in `docs/plans/task.md` throughout execution.

---

### Task 1: Setup Task Tracking & Git Branch

**Files:**
- Create: `docs/plans/task.md`

**Interfaces:**
- Consumes: `main`
- Produces: Branch `feature/002-domain-model-structs` and live progress tracker in `docs/plans/task.md`

- [x] **Step 1: Create live progress tracker in `docs/plans/task.md`**

Create `docs/plans/task.md` with table format:

```markdown
# Live Task Tracker: 002 Domain Model Structs

| Task # | Task Description | Status | Verification |
|---|---|---|---|
| Task 1 | Setup Task Tracking & Git Branch | In Progress | git branch checked |
| Task 2 | Enums (`ArrowStatus`) | Pending | `go test ./internal/model/... -run TestEnums_JSON` |
| Task 3 | Bow Domain Models (`bow.go`) | Pending | `go test ./internal/model/... -run TestBowModels_JSON` |
| Task 4 | Arrow Domain Models (`arrow.go`) | Pending | `go test ./internal/model/... -run TestArrowModels_JSON` |
| Task 5 | Refactor Archer Models (`archer.go`) | Pending | `go test ./internal/model/... -run TestArcherModels_JSON` |
| Task 6 | Refactor Auth Models (`auth.go`) | Pending | `go test ./internal/model/... -run TestAuthModels_JSON` |
| Task 7 | Package Verification & Quality Assurance | Pending | `go test`, `go vet`, `golangci-lint` |
| Task 8 | Mark Task 002 Documentation & Live Tracker as Completed | Pending | checklist review |
```

- [x] **Step 2: Create and checkout git branch `feature/002-domain-model-structs`**

Run:
```bash
git checkout -b feature/002-domain-model-structs
```
Verify:
```bash
git branch --show-current
```
Expected output: `feature/002-domain-model-structs`

- [x] **Step 3: Update `docs/plans/task.md`**

Update status of Task 1 to `Done`.

---

### Task 2: Enums (`ArrowStatus`)

**Files:**
- Modify: `backend/internal/model/enums.go:60-70`
- Test: `backend/internal/model/model_test.go:12-49`

**Interfaces:**
- Consumes: none
- Produces:
  - `ArrowStatus` enum type (`string`)
  - `ArrowStatusInUse ArrowStatus = "in_use"`
  - `ArrowStatusDamaged ArrowStatus = "damaged"`
  - `ArrowStatusLost ArrowStatus = "lost"`
  - Compatibility alias `ArrowStatusType = ArrowStatus`

- [x] **Step 1: Write failing test in `model_test.go`**

Add test cases for `ArrowStatus` values to `TestEnums_JSON` in `backend/internal/model/model_test.go`:

```go
		{"ArrowStatusInUse", model.ArrowStatusInUse, `"in_use"`},
		{"ArrowStatusDamaged", model.ArrowStatusDamaged, `"damaged"`},
		{"ArrowStatusLost", model.ArrowStatusLost, `"lost"`},
```

- [x] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestEnums_JSON
```
Expected output: Compilation failure (`undefined: model.ArrowStatusInUse`).

- [x] **Step 3: Implement `ArrowStatus` enum in `enums.go`**

Add to `backend/internal/model/enums.go`:

```go
// ArrowStatus represents the operational status of an arrow.
type ArrowStatus string

// ArrowStatusType is a compatibility alias for frontend OpenAPI schema generation.
type ArrowStatusType = ArrowStatus

const (
	ArrowStatusInUse   ArrowStatus = "in_use"
	ArrowStatusDamaged ArrowStatus = "damaged"
	ArrowStatusLost    ArrowStatus = "lost"
)
```

- [x] **Step 4: Run test to verify it passes**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestEnums_JSON
```
Expected output: `PASS: TestEnums_JSON` (with subtests for `ArrowStatusInUse`, `ArrowStatusDamaged`, `ArrowStatusLost`).

- [x] **Step 5: Commit changes**

Run:
```bash
git add backend/internal/model/enums.go backend/internal/model/model_test.go
git commit -m "feat(model): add ArrowStatus enum"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 2 to `Done`.

---

### Task 3: Bow Domain Models (`bow.go`)

**Files:**
- Create: `backend/internal/model/bow.go`
- Test: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: `model.Bowstyle`, `uuid.UUID`, `time.Time`
- Produces:
  - `BowCreate`: `ArcherID uuid.UUID`, `Name string`, `Bowstyle Bowstyle`, `DrawWeight float64`
  - `BowRead`: `BowID uuid.UUID`, `ArcherID uuid.UUID`, `Name string`, `Bowstyle Bowstyle`, `DrawWeight float64`, `IsDeleted bool`, `CreatedAt time.Time`
  - `BowSet`: `Name *string`, `Bowstyle *Bowstyle`, `DrawWeight *float64`
  - `BowFilter`: `BowID *uuid.UUID`, `ArcherID *uuid.UUID`, `Bowstyle *Bowstyle`, `IsDeleted *bool`
  - `BowUpdate`: `Where BowFilter`, `Data BowSet`

- [x] **Step 1: Write failing test in `model_test.go`**

Add `TestBowModels_JSON` to `backend/internal/model/model_test.go`:

```go
func TestBowModels_JSON(t *testing.T) {
	archerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bowID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("BowCreate JSON marshaling", func(t *testing.T) {
		create := model.BowCreate{
			ArcherID:   archerID,
			Name:       "Competition Recurve",
			Bowstyle:   model.BowstyleRecurve,
			DrawWeight: 38.5,
		}

		b, err := json.Marshal(create)
		if err != nil {
			t.Fatalf("failed to marshal BowCreate: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal BowCreate JSON: %v", err)
		}

		expectedKeys := []string{"archer_id", "name", "bowstyle", "draw_weight"}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected JSON key %q missing in BowCreate", k)
			}
		}

		var decoded model.BowCreate
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to roundtrip unmarshal BowCreate: %v", err)
		}
		if decoded.Name != "Competition Recurve" || decoded.DrawWeight != 38.5 || decoded.Bowstyle != model.BowstyleRecurve {
			t.Errorf("mismatch in decoded BowCreate: %+v", decoded)
		}
	})

	t.Run("BowRead JSON marshaling", func(t *testing.T) {
		read := model.BowRead{
			BowID:      bowID,
			ArcherID:   archerID,
			Name:       "Competition Recurve",
			Bowstyle:   model.BowstyleRecurve,
			DrawWeight: 38.5,
			IsDeleted:  false,
			CreatedAt:  now,
		}

		b, err := json.Marshal(read)
		if err != nil {
			t.Fatalf("failed to marshal BowRead: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal BowRead JSON: %v", err)
		}

		expectedKeys := []string{"bow_id", "archer_id", "name", "bowstyle", "draw_weight", "is_deleted", "created_at"}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected JSON key %q missing in BowRead", k)
			}
		}

		var decoded model.BowRead
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to roundtrip unmarshal BowRead: %v", err)
		}
		if decoded.BowID != bowID || decoded.IsDeleted != false || !decoded.CreatedAt.Equal(now) {
			t.Errorf("mismatch in decoded BowRead: %+v", decoded)
		}
	})

	t.Run("BowSet and BowFilter JSON marshaling", func(t *testing.T) {
		name := "Updated Bow Name"
		weight := 42.0
		set := model.BowSet{
			Name:       &name,
			DrawWeight: &weight,
		}

		sb, err := json.Marshal(set)
		if err != nil {
			t.Fatalf("failed to marshal BowSet: %v", err)
		}
		var setMap map[string]any
		if err := json.Unmarshal(sb, &setMap); err != nil {
			t.Fatalf("failed to unmarshal BowSet JSON: %v", err)
		}
		if _, ok := setMap["name"]; !ok {
			t.Errorf("expected 'name' in BowSet")
		}
		if _, ok := setMap["bowstyle"]; ok {
			t.Errorf("expected omitted 'bowstyle' in BowSet")
		}

		isDeleted := false
		filter := model.BowFilter{
			ArcherID:  &archerID,
			IsDeleted: &isDeleted,
		}
		fb, err := json.Marshal(filter)
		if err != nil {
			t.Fatalf("failed to marshal BowFilter: %v", err)
		}
		var filterMap map[string]any
		if err := json.Unmarshal(fb, &filterMap); err != nil {
			t.Fatalf("failed to unmarshal BowFilter JSON: %v", err)
		}
		if _, ok := filterMap["archer_id"]; !ok {
			t.Errorf("expected 'archer_id' in BowFilter")
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestBowModels_JSON
```
Expected output: Compilation failure (`undefined: model.BowCreate`).

- [x] **Step 3: Implement `backend/internal/model/bow.go`**

Create `backend/internal/model/bow.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// BowCreate represents the payload required to register a new bow.
type BowCreate struct {
	ArcherID   uuid.UUID `json:"archer_id"`
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	Bowstyle   Bowstyle  `json:"bowstyle" validate:"required"`
	DrawWeight float64   `json:"draw_weight" validate:"required,gt=0,lte=200"`
}

// BowRead represents the full persisted bow domain model.
type BowRead struct {
	BowID      uuid.UUID `json:"bow_id"`
	ArcherID   uuid.UUID `json:"archer_id"`
	Name       string    `json:"name"`
	Bowstyle   Bowstyle  `json:"bowstyle"`
	DrawWeight float64   `json:"draw_weight"`
	IsDeleted  bool      `json:"is_deleted"`
	CreatedAt  time.Time `json:"created_at"`
}

// BowSet represents mutable fields when updating bow specifications.
type BowSet struct {
	Name       *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Bowstyle   *Bowstyle `json:"bowstyle,omitempty"`
	DrawWeight *float64  `json:"draw_weight,omitempty" validate:"omitempty,gt=0,lte=200"`
}

// BowFilter represents criteria to query or filter bows.
type BowFilter struct {
	BowID     *uuid.UUID `json:"bow_id,omitempty"`
	ArcherID  *uuid.UUID `json:"archer_id,omitempty"`
	Bowstyle  *Bowstyle  `json:"bowstyle,omitempty"`
	IsDeleted *bool      `json:"is_deleted,omitempty"`
}

// BowUpdate wraps target filter criteria and field updates for bows.
type BowUpdate struct {
	Where BowFilter `json:"where"`
	Data  BowSet    `json:"data"`
}
```

- [x] **Step 4: Run test to verify it passes**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestBowModels_JSON
```
Expected output: `PASS: TestBowModels_JSON`.

- [x] **Step 5: Commit changes**

Run:
```bash
git add backend/internal/model/bow.go backend/internal/model/model_test.go
git commit -m "feat(model): implement bow domain model structs"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 3 to `Done`.

---

### Task 4: Arrow Domain Models (`arrow.go`)

**Files:**
- Create: `backend/internal/model/arrow.go`
- Test: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: `model.ArrowStatus`, `uuid.UUID`, `time.Time`
- Produces:
  - `ArrowCreate`: `ArcherID uuid.UUID`, `ArrowSet int16`, `ArrowNumber int16`, `Status *ArrowStatus`, `Spine *float64`, `Length *float64`, `Weight *float64`
  - `ArrowBatchCreate`: `ArcherID uuid.UUID`, `ArrowSet int16`, `Count int`, `Status *ArrowStatus`, `Spine *float64`, `Length *float64`, `Weight *float64`
  - `ArrowRead`: `ArrowID uuid.UUID`, `ArcherID uuid.UUID`, `ArrowSet int16`, `ArrowNumber int16`, `Status ArrowStatus`, `Spine *float64`, `Length *float64`, `Weight *float64`, `IsDeleted bool`, `CreatedAt time.Time`
  - `ArrowSet`: `Status *ArrowStatus`, `Spine *float64`, `Length *float64`, `Weight *float64`
  - `ArrowFilter`: `ArrowID *uuid.UUID`, `ArcherID *uuid.UUID`, `ArrowSet *int16`, `Status *ArrowStatus`, `IsDeleted *bool`
  - `ArrowUpdate`: `Where ArrowFilter`, `Data ArrowSet`

- [x] **Step 1: Write failing test in `model_test.go`**

Add `TestArrowModels_JSON` to `backend/internal/model/model_test.go`:

```go
func TestArrowModels_JSON(t *testing.T) {
	archerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	arrowID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	now := time.Now().UTC().Truncate(time.Second)
	spine := 500.0
	length := 29.5
	weight := 320.0
	status := model.ArrowStatusInUse

	t.Run("ArrowCreate JSON marshaling", func(t *testing.T) {
		create := model.ArrowCreate{
			ArcherID:    archerID,
			ArrowSet:    1,
			ArrowNumber: 4,
			Status:      &status,
			Spine:       &spine,
			Length:      &length,
			Weight:      &weight,
		}

		b, err := json.Marshal(create)
		if err != nil {
			t.Fatalf("failed to marshal ArrowCreate: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal ArrowCreate JSON: %v", err)
		}

		expectedKeys := []string{"archer_id", "arrow_set", "arrow_number", "status", "spine", "length", "weight"}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected JSON key %q missing in ArrowCreate", k)
			}
		}

		var decoded model.ArrowCreate
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to roundtrip unmarshal ArrowCreate: %v", err)
		}
		if decoded.ArrowSet != 1 || decoded.ArrowNumber != 4 || *decoded.Spine != 500.0 {
			t.Errorf("mismatch in decoded ArrowCreate: %+v", decoded)
		}
	})

	t.Run("ArrowBatchCreate JSON marshaling", func(t *testing.T) {
		batch := model.ArrowBatchCreate{
			ArcherID: archerID,
			ArrowSet: 2,
			Count:    6,
			Status:   &status,
			Spine:    &spine,
			Length:   &length,
			Weight:   &weight,
		}

		b, err := json.Marshal(batch)
		if err != nil {
			t.Fatalf("failed to marshal ArrowBatchCreate: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal ArrowBatchCreate JSON: %v", err)
		}

		expectedKeys := []string{"archer_id", "arrow_set", "count", "status", "spine", "length", "weight"}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected JSON key %q missing in ArrowBatchCreate", k)
			}
		}

		var decoded model.ArrowBatchCreate
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to roundtrip unmarshal ArrowBatchCreate: %v", err)
		}
		if decoded.Count != 6 || decoded.ArrowSet != 2 {
			t.Errorf("mismatch in decoded ArrowBatchCreate: %+v", decoded)
		}
	})

	t.Run("ArrowRead JSON marshaling", func(t *testing.T) {
		read := model.ArrowRead{
			ArrowID:     arrowID,
			ArcherID:    archerID,
			ArrowSet:    1,
			ArrowNumber: 3,
			Status:      model.ArrowStatusInUse,
			Spine:       &spine,
			Length:      &length,
			Weight:      &weight,
			IsDeleted:   false,
			CreatedAt:   now,
		}

		b, err := json.Marshal(read)
		if err != nil {
			t.Fatalf("failed to marshal ArrowRead: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal ArrowRead JSON: %v", err)
		}

		expectedKeys := []string{
			"arrow_id", "archer_id", "arrow_set", "arrow_number",
			"status", "spine", "length", "weight", "is_deleted", "created_at",
		}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected JSON key %q missing in ArrowRead", k)
			}
		}

		var decoded model.ArrowRead
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to roundtrip unmarshal ArrowRead: %v", err)
		}
		if decoded.ArrowID != arrowID || decoded.Status != model.ArrowStatusInUse || decoded.IsDeleted != false {
			t.Errorf("mismatch in decoded ArrowRead: %+v", decoded)
		}
	})

	t.Run("ArrowSet and ArrowFilter JSON marshaling", func(t *testing.T) {
		damaged := model.ArrowStatusDamaged
		set := model.ArrowSet{
			Status: &damaged,
		}
		sb, err := json.Marshal(set)
		if err != nil {
			t.Fatalf("failed to marshal ArrowSet: %v", err)
		}
		var setMap map[string]any
		if err := json.Unmarshal(sb, &setMap); err != nil {
			t.Fatalf("failed to unmarshal ArrowSet JSON: %v", err)
		}
		if _, ok := setMap["status"]; !ok {
			t.Errorf("expected 'status' in ArrowSet")
		}
		if _, ok := setMap["spine"]; ok {
			t.Errorf("expected omitted 'spine' in ArrowSet")
		}

		setNum := int16(1)
		isDeleted := false
		filter := model.ArrowFilter{
			ArcherID:  &archerID,
			ArrowSet:  &setNum,
			Status:    &damaged,
			IsDeleted: &isDeleted,
		}
		fb, err := json.Marshal(filter)
		if err != nil {
			t.Fatalf("failed to marshal ArrowFilter: %v", err)
		}
		var filterMap map[string]any
		if err := json.Unmarshal(fb, &filterMap); err != nil {
			t.Fatalf("failed to unmarshal ArrowFilter JSON: %v", err)
		}
		if _, ok := filterMap["arrow_set"]; !ok {
			t.Errorf("expected 'arrow_set' in ArrowFilter")
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestArrowModels_JSON
```
Expected output: Compilation failure (`undefined: model.ArrowCreate`).

- [x] **Step 3: Implement `backend/internal/model/arrow.go`**

Create `backend/internal/model/arrow.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// ArrowCreate represents the payload required to register an individual arrow.
type ArrowCreate struct {
	ArcherID    uuid.UUID    `json:"archer_id"`
	ArrowSet    int16        `json:"arrow_set" validate:"required,gt=0"`
	ArrowNumber int16        `json:"arrow_number" validate:"required,gt=0"`
	Status      *ArrowStatus `json:"status,omitempty"`
	Spine       *float64     `json:"spine,omitempty" validate:"omitempty,gt=0"`
	Length      *float64     `json:"length,omitempty" validate:"omitempty,gt=0"`
	Weight      *float64     `json:"weight,omitempty" validate:"omitempty,gt=0"`
}

// ArrowBatchCreate represents the payload required to register an entire batch of arrows.
type ArrowBatchCreate struct {
	ArcherID uuid.UUID    `json:"archer_id"`
	ArrowSet int16        `json:"arrow_set" validate:"required,gt=0"`
	Count    int          `json:"count" validate:"required,gte=3"`
	Status   *ArrowStatus `json:"status,omitempty"`
	Spine    *float64     `json:"spine,omitempty" validate:"omitempty,gt=0"`
	Length   *float64     `json:"length,omitempty" validate:"omitempty,gt=0"`
	Weight   *float64     `json:"weight,omitempty" validate:"omitempty,gt=0"`
}

// ArrowRead represents the full persisted arrow domain model.
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

// ArrowSet represents mutable fields when updating arrow specifications or status.
type ArrowSet struct {
	Status *ArrowStatus `json:"status,omitempty"`
	Spine  *float64     `json:"spine,omitempty" validate:"omitempty,gt=0"`
	Length *float64     `json:"length,omitempty" validate:"omitempty,gt=0"`
	Weight *float64     `json:"weight,omitempty" validate:"omitempty,gt=0"`
}

// ArrowFilter represents criteria to query or filter arrows.
type ArrowFilter struct {
	ArrowID   *uuid.UUID   `json:"arrow_id,omitempty"`
	ArcherID  *uuid.UUID   `json:"archer_id,omitempty"`
	ArrowSet  *int16       `json:"arrow_set,omitempty"`
	Status    *ArrowStatus `json:"status,omitempty"`
	IsDeleted *bool        `json:"is_deleted,omitempty"`
}

// ArrowUpdate wraps target filter criteria and field updates for arrows.
type ArrowUpdate struct {
	Where ArrowFilter `json:"where"`
	Data  ArrowSet    `json:"data"`
}
```

- [x] **Step 4: Run test to verify it passes**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestArrowModels_JSON
```
Expected output: `PASS: TestArrowModels_JSON`.

- [x] **Step 5: Commit changes**

Run:
```bash
git add backend/internal/model/arrow.go backend/internal/model/model_test.go
git commit -m "feat(model): implement arrow domain model structs"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 4 to `Done`.

---

### Task 5: Refactor Archer Models (`archer.go`)

**Files:**
- Modify: `backend/internal/model/archer.go`
- Test: `backend/internal/model/model_test.go:91-160`

**Interfaces:**
- Consumes: `model.Gender`, `uuid.UUID`
- Produces:
  - `ArcherCreate`: `ArcherID *uuid.UUID`, `Email string`, `FirstName string`, `LastName string`, `DateOfBirth string`, `Gender Gender`
  - `ArcherRead`: `ArcherID uuid.UUID`, `Email string`, `FirstName string`, `LastName string`, `DateOfBirth string`, `Gender Gender`, `IsDeleted bool`
  - `ArcherSet`: `Email *string`, `FirstName *string`, `LastName *string`, `DateOfBirth *string`, `Gender *Gender`
  - `ArcherFilter`: `ArcherID *uuid.UUID`, `Email *string`, `FirstName *string`, `LastName *string`, `Gender *Gender`, `IsDeleted *bool`
  - `ArcherUpdate`: `Where ArcherFilter`, `Data ArcherSet`
  - `ArcherID`: `ArcherID uuid.UUID`

- [x] **Step 1: Update `TestArcherModels_JSON` in `model_test.go`**

Update `TestArcherModels_JSON` in `backend/internal/model/model_test.go`:

```go
func TestArcherModels_JSON(t *testing.T) {
	create := model.ArcherCreate{
		FirstName:   "Robin",
		LastName:    "Hood",
		Email:       "robin@sherwood.org",
		DateOfBirth: "1995-06-15",
		Gender:      model.GenderMale,
	}

	b, err := json.Marshal(create)
	if err != nil {
		t.Fatalf("failed to marshal ArcherCreate: %v", err)
	}

	var createMap map[string]any
	if err := json.Unmarshal(b, &createMap); err != nil {
		t.Fatalf("failed to unmarshal ArcherCreate JSON: %v", err)
	}

	expectedKeys := []string{
		"first_name", "last_name", "email", "date_of_birth", "gender",
	}
	for _, k := range expectedKeys {
		if _, ok := createMap[k]; !ok {
			t.Errorf("expected JSON key %q missing in ArcherCreate serialization", k)
		}
	}

	forbiddenKeys := []string{
		"bowstyle", "draw_weight", "club_id", "google_picture_url", "google_subject", "last_login_at", "created_at",
	}
	for _, k := range forbiddenKeys {
		if _, ok := createMap[k]; ok {
			t.Errorf("forbidden legacy key %q found in ArcherCreate serialization", k)
		}
	}

	archerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	read := model.ArcherRead{
		ArcherID:    archerID,
		FirstName:   create.FirstName,
		LastName:    create.LastName,
		Email:       create.Email,
		DateOfBirth: create.DateOfBirth,
		Gender:      create.Gender,
		IsDeleted:   false,
	}

	rb, err := json.Marshal(read)
	if err != nil {
		t.Fatalf("failed to marshal ArcherRead: %v", err)
	}

	var readMap map[string]any
	if err := json.Unmarshal(rb, &readMap); err != nil {
		t.Fatalf("failed to unmarshal ArcherRead JSON: %v", err)
	}

	expectedReadKeys := []string{
		"archer_id", "first_name", "last_name", "email", "date_of_birth", "gender", "is_deleted",
	}
	for _, k := range expectedReadKeys {
		if _, ok := readMap[k]; !ok {
			t.Errorf("expected JSON key %q missing in ArcherRead serialization", k)
		}
	}

	for _, k := range forbiddenKeys {
		if _, ok := readMap[k]; ok {
			t.Errorf("forbidden legacy key %q found in ArcherRead serialization", k)
		}
	}

	var decodedRead model.ArcherRead
	if err := json.Unmarshal(rb, &decodedRead); err != nil {
		t.Fatalf("failed to unmarshal ArcherRead: %v", err)
	}

	if decodedRead.ArcherID != archerID || decodedRead.Email != create.Email || decodedRead.Gender != model.GenderMale || decodedRead.IsDeleted != false {
		t.Errorf("mismatch in decoded ArcherRead: %+v", decodedRead)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestArcherModels_JSON
```
Expected output: Test failure on forbidden keys or struct fields.

- [x] **Step 3: Refactor `backend/internal/model/archer.go`**

Replace `backend/internal/model/archer.go` with personal-profile-only domain models:

```go
package model

import (
	"github.com/google/uuid"
)

// ArcherCreate represents the payload required to create a new archer profile.
type ArcherCreate struct {
	ArcherID    *uuid.UUID `json:"archer_id,omitempty"`
	FirstName   string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName    string     `json:"last_name" validate:"required,min=1,max=100"`
	Email       string     `json:"email" validate:"required,email"`
	DateOfBirth string     `json:"date_of_birth" validate:"required"`
	Gender      Gender     `json:"gender" validate:"required"`
}

// ArcherSet represents mutable fields when updating an archer profile.
type ArcherSet struct {
	FirstName   *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName    *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Email       *string `json:"email,omitempty" validate:"omitempty,email"`
	DateOfBirth *string `json:"date_of_birth,omitempty"`
	Gender      *Gender `json:"gender,omitempty"`
}

// ArcherFilter represents criteria to query or select archers.
type ArcherFilter struct {
	ArcherID  *uuid.UUID `json:"archer_id,omitempty"`
	Email     *string    `json:"email,omitempty"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Gender    *Gender    `json:"gender,omitempty"`
	IsDeleted *bool      `json:"is_deleted,omitempty"`
}

// ArcherUpdate wraps target filter criteria and field updates for archers.
type ArcherUpdate struct {
	Where ArcherFilter `json:"where"`
	Data  ArcherSet    `json:"data"`
}

// ArcherRead represents the full persisted archer personal profile domain model.
type ArcherRead struct {
	ArcherID    uuid.UUID `json:"archer_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      Gender    `json:"gender"`
	IsDeleted   bool      `json:"is_deleted"`
}

// ArcherID represents a standalone archer identifier wrapper.
type ArcherID struct {
	ArcherID uuid.UUID `json:"archer_id"`
}
```

- [x] **Step 4: Run test to verify it passes**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestArcherModels_JSON
```
Expected output: `PASS: TestArcherModels_JSON`.

- [x] **Step 5: Commit changes**

Run:
```bash
git add backend/internal/model/archer.go backend/internal/model/model_test.go
git commit -m "refactor(model): simplify archer models to personal profile data"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 5 to `Done`.

---

### Task 6: Refactor Auth Models (`auth.go`)

**Files:**
- Modify: `backend/internal/model/auth.go`
- Test: `backend/internal/model/model_test.go:161-198`

**Interfaces:**
- Consumes: `model.ArcherRead`, `model.AuthStatus`, `uuid.UUID`, `time.Time`
- Produces:
  - `AuthIdentityRead`: `ArcherID uuid.UUID`, `GoogleSubject string`, `GooglePictureURL *string`, `LastLoginAt time.Time`, `CreatedAt time.Time`
  - Updated `TestAuthModels_JSON` covering `AuthAuthenticated`, `AuthNeedsRegistration`, `AuthIdentityRead`

- [x] **Step 1: Write failing test in `model_test.go`**

Update `TestAuthModels_JSON` in `backend/internal/model/model_test.go` to test `AuthIdentityRead` and updated `ArcherRead` in `AuthAuthenticated`:

```go
func TestAuthModels_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	archerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	picURL := "https://example.com/avatar.jpg"

	t.Run("AuthIdentityRead JSON marshaling", func(t *testing.T) {
		identity := model.AuthIdentityRead{
			ArcherID:         archerID,
			GoogleSubject:    "google-sub-98765",
			GooglePictureURL: &picURL,
			LastLoginAt:      now,
			CreatedAt:        now,
		}

		b, err := json.Marshal(identity)
		if err != nil {
			t.Fatalf("failed to marshal AuthIdentityRead: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal AuthIdentityRead JSON: %v", err)
		}

		expectedKeys := []string{
			"archer_id", "google_subject", "google_picture_url", "last_login_at", "created_at",
		}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected key %q in AuthIdentityRead JSON", k)
			}
		}

		var decoded model.AuthIdentityRead
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to unmarshal AuthIdentityRead: %v", err)
		}
		if decoded.ArcherID != archerID || decoded.GoogleSubject != "google-sub-98765" || *decoded.GooglePictureURL != picURL {
			t.Errorf("mismatch in decoded AuthIdentityRead: %+v", decoded)
		}
	})

	t.Run("AuthAuthenticated JSON marshaling", func(t *testing.T) {
		authRead := model.AuthAuthenticated{
			Status:      model.AuthStatusAuthenticated,
			AccessToken: "jwt-token-xyz",
			ExpiresAt:   now.Add(time.Hour * 24),
			Archer: model.ArcherRead{
				ArcherID:    archerID,
				FirstName:   "Robin",
				LastName:    "Hood",
				Email:       "robin@sherwood.org",
				DateOfBirth: "1995-06-15",
				Gender:      model.GenderMale,
				IsDeleted:   false,
			},
		}

		b, err := json.Marshal(authRead)
		if err != nil {
			t.Fatalf("failed to marshal AuthAuthenticated: %v", err)
		}

		var decoded model.AuthAuthenticated
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to unmarshal AuthAuthenticated: %v", err)
		}

		if decoded.Status != model.AuthStatusAuthenticated || decoded.AccessToken != "jwt-token-xyz" || decoded.Archer.IsDeleted != false {
			t.Errorf("mismatch in decoded AuthAuthenticated: %+v", decoded)
		}
	})

	t.Run("AuthNeedsRegistration JSON marshaling", func(t *testing.T) {
		given := "Robin"
		family := "Hood"
		needsReg := model.AuthNeedsRegistration{
			Status:             model.AuthStatusNeedsRegistration,
			GoogleEmail:        "robin@sherwood.org",
			GoogleSubject:      "google-sub-12345",
			GivenName:          &given,
			FamilyName:         &family,
			GivenNameProvided:  true,
			FamilyNameProvided: true,
			PictureURL:         &picURL,
		}

		b, err := json.Marshal(needsReg)
		if err != nil {
			t.Fatalf("failed to marshal AuthNeedsRegistration: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("failed to unmarshal AuthNeedsRegistration JSON: %v", err)
		}

		expectedKeys := []string{
			"status", "google_email", "google_subject", "given_name", "family_name",
			"given_name_provided", "family_name_provided", "picture_url",
		}
		for _, k := range expectedKeys {
			if _, ok := m[k]; !ok {
				t.Errorf("expected key %q in AuthNeedsRegistration JSON", k)
			}
		}

		var decoded model.AuthNeedsRegistration
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to unmarshal AuthNeedsRegistration: %v", err)
		}
		if decoded.GoogleEmail != "robin@sherwood.org" || !decoded.GivenNameProvided {
			t.Errorf("mismatch in decoded AuthNeedsRegistration: %+v", decoded)
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestAuthModels_JSON
```
Expected output: Compilation failure (`undefined: model.AuthIdentityRead`).

- [x] **Step 3: Define `AuthIdentityRead` in `auth.go`**

Add `AuthIdentityRead` to `backend/internal/model/auth.go`:

```go
// AuthIdentityRead represents a persisted Google OAuth authentication record.
type AuthIdentityRead struct {
	ArcherID         uuid.UUID `json:"archer_id"`
	GoogleSubject    string    `json:"google_subject"`
	GooglePictureURL *string   `json:"google_picture_url,omitempty"`
	LastLoginAt      time.Time `json:"last_login_at"`
	CreatedAt        time.Time `json:"created_at"`
}
```

- [x] **Step 4: Run test to verify it passes**

Run:
```bash
cd backend && go test ./internal/model/... -v -run TestAuthModels_JSON
```
Expected output: `PASS: TestAuthModels_JSON`.

- [x] **Step 5: Commit changes**

Run:
```bash
git add backend/internal/model/auth.go backend/internal/model/model_test.go
git commit -m "refactor(model): add AuthIdentityRead and update auth models"
```

- [x] **Step 6: Update `docs/plans/task.md`**

Update status of Task 6 to `Done`.

---

### Task 7: Package Verification & Quality Assurance

**Files:**
- Modify: `backend/internal/model/...` (formatting only if needed)

**Interfaces:**
- Consumes: All models in `backend/internal/model`
- Produces: Clean test and lint runs across `backend/internal/model`

- [x] **Step 1: Run the full model package test suite**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected output: All unit tests in `internal/model` PASS with code 0.

- [x] **Step 2: Run `go vet` on the model package**

Run:
```bash
cd backend && go vet ./internal/model/...
```
Expected output: Exit code 0, no issues reported.

- [x] **Step 3: Run `golangci-lint` on the model package**

Run:
```bash
cd backend && golangci-lint run ./internal/model/...
```
Expected output: `0 issues` with exit code 0.

- [x] **Step 4: Run `gofumpt` formatting check**

Run:
```bash
cd backend && gofumpt -l -w internal/model/
```
Verify:
```bash
git diff --name-only backend/internal/model/
```
If formatting changes were made:
```bash
git add backend/internal/model
git commit -m "style(model): apply gofumpt formatting"
```

- [x] **Step 5: Update `docs/plans/task.md`**

Update status of Task 7 to `Done`.

---

### Task 8: Mark Task 002 Documentation & Live Tracker as Completed

**Files:**
- Modify: `docs/onboarding_refactor/tasks/002_domain_model_structs.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified code in `backend/internal/model`
- Produces: Updated checklist in `docs/onboarding_refactor/tasks/002_domain_model_structs.md` and complete live tracker

- [x] **Step 1: Check off all acceptance criteria and steps in `docs/onboarding_refactor/tasks/002_domain_model_structs.md`**

Mark all items in `002_domain_model_structs.md` as done:

```markdown
## Acceptance Criteria

- [x] New file `backend/internal/model/bow.go` defines:
    - [x] `BowCreate`: `ArcherID`, `Name`, `Bowstyle`, `DrawWeight`
    - [x] `BowRead`: `BowID`, `ArcherID`, `Name`, `Bowstyle`, `DrawWeight`, `IsDeleted`, `CreatedAt`
    - [x] `BowSet`: nullable fields for updates (`Name`, `Bowstyle`, `DrawWeight`)
    - [x] `BowFilter`: `BowID`, `ArcherID`, `Bowstyle`, `IsDeleted`
- [x] New enum `ArrowStatus` in `backend/internal/model/enums.go`:
    - [x] Values: `ArrowStatusInUse = "in_use"`, `ArrowStatusDamaged = "damaged"`, `ArrowStatusLost = "lost"`
- [x] New file `backend/internal/model/arrow.go` defines:
    - [x] `ArrowCreate`: `ArcherID`, `ArrowSet`, `ArrowNumber`, `Status` (optional, defaults to `in_use`), `Spine`, `Length`, `Weight`
    - [x] `ArrowBatchCreate`: `ArcherID`, `ArrowSet`, `Count` (min 3), `Status` (optional, defaults to `in_use`), `Spine`, `Length`, `Weight`
    - [x] `ArrowRead`: `ArrowID`, `ArcherID`, `ArrowSet`, `ArrowNumber`, `Status`, `Spine`, `Length`, `Weight`, `IsDeleted`, `CreatedAt`
    - [x] `ArrowSet`: nullable fields for updates (`Status`, `Spine`, `Length`, `Weight`)
    - [x] `ArrowFilter`: `ArrowID`, `ArcherID`, `ArrowSet`, `Status`, `IsDeleted`
- [x] Modified `backend/internal/model/archer.go`:
    - [x] Equipment fields removed (`Bowstyle`, `DrawWeight`)
    - [x] Legacy fields removed (`ClubID`, `GoogleSubject`, `GooglePictureURL`, `LastLoginAt`, `CreatedAt`)
    - [x] `ArcherCreate` requires: `Email`, `FirstName`, `LastName`, `DateOfBirth`, `Gender`
    - [x] `ArcherRead` includes: `ArcherID`, `Email`, `FirstName`, `LastName`, `DateOfBirth`, `Gender`, `IsDeleted`
- [x] Modified `backend/internal/model/auth.go`:
    - [x] `AuthIdentityRead` represents OAuth record: `ArcherID`, `GoogleSubject`, `GooglePictureURL`, `LastLoginAt`, `CreatedAt`
    - [x] `AuthStatus` enum values preserved: `authenticated`, `needs_registration`
    - [x] `AuthNeedsRegistration` includes `GoogleEmail`, `GoogleSubject`, `GivenName`, `FamilyName`, `PictureURL`
- [x] JSON tags match snake_case API specifications.
- [x] Unit tests in `backend/internal/model/model_test.go` verify JSON marshaling and validation.
- [x] `cd backend && go test ./internal/model/... -v` passes.
- [x] `cd backend && go vet ./internal/model/...` reports no issues.
```

And in `## Steps`:
```markdown
- [x] **Step 1: Write failing tests in `model_test.go`**
- [x] **Step 2: Run tests to verify they fail**
- [x] **Step 3: Implement `bow.go`**
- [x] **Step 4: Implement `arrow.go`**
- [x] **Step 5: Refactor `archer.go` and `auth.go`**
- [x] **Step 6: Run tests to verify they pass**
- [x] **Step 7: Commit changes**
```

- [x] **Step 2: Finalize `docs/plans/task.md`**

Update `docs/plans/task.md` with all tasks marked `Done`:

```markdown
# Live Task Tracker: 002 Domain Model Structs

| Task # | Task Description | Status | Verification |
|---|---|---|---|
| Task 1 | Setup Task Tracking & Git Branch | Done | branch `feature/002-domain-model-structs` active |
| Task 2 | Enums (`ArrowStatus`) | Done | `TestEnums_JSON` passed |
| Task 3 | Bow Domain Models (`bow.go`) | Done | `TestBowModels_JSON` passed |
| Task 4 | Arrow Domain Models (`arrow.go`) | Done | `TestArrowModels_JSON` passed |
| Task 5 | Refactor Archer Models (`archer.go`) | Done | `TestArcherModels_JSON` passed |
| Task 6 | Refactor Auth Models (`auth.go`) | Done | `TestAuthModels_JSON` passed |
| Task 7 | Package Verification & Quality Assurance | Done | `go test`, `go vet`, `golangci-lint` clean |
| Task 8 | Mark Task 002 Documentation & Live Tracker as Completed | Done | all checklists updated |
```

- [x] **Step 3: Commit documentation updates**

Run:
```bash
git add docs/onboarding_refactor/tasks/002_domain_model_structs.md docs/plans/task.md
git commit -m "docs(tasks): mark 002 domain model structs tasks as completed"
```
