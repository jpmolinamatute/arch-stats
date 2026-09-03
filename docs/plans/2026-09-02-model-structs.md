# Domain Model Structs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define all domain model structs in `internal/model/` with JSON tags, validation tags, and typed string enums matching the existing Pydantic schemas and OpenAPI specifications.

**Architecture:** Pure Go data structures defined across domain-focused files (`enums.go`, `base.go`, `archer.go`, `auth.go`, `session.go`, `slot.go`, `shot.go`, `face.go`, `target.go`, `live_stats.go`, and `report.go`). All types adhere to the single-responsibility principle with explicit snake_case JSON tags, pointer types (`*T`) for nullable/optional fields, `uuid.UUID` for identifiers, and `time.Time` for timestamps.

**Tech Stack:** Go 1.27+, `github.com/google/uuid`, standard library (`time`, `encoding/json`).

**Spec:** [docs/go_refactor/tasks/007-model_structs.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/007-model_structs.md)

## Global Constraints

- Git branch: `refactor/007-model-structs`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/model`
- All structs must have exact snake_case JSON tags matching the API contract.
- Enums are implemented as typed string constants.
- Optional and nullable fields must use pointer types (`*T`) with `omitempty` JSON tag.
- `uuid.UUID` (from `github.com/google/uuid`) must be used for ID fields.
- `time.Time` must be used for datetime fields; `string` ("YYYY-MM-DD") for date-only fields (`date_of_birth`).
- `go test -race ./internal/model/... -v` must pass.
- `go vet ./...` must report no issues.
- `golangci-lint run ./...` must report no issues.
- `go build ./...` must compile without errors.

---

## File Structure

```
backend/
├── go.mod                                   # Add direct dependency github.com/google/uuid
├── go.sum
└── internal/
    └── model/
        ├── .gitkeep                         # [DELETE] Remove once package files exist
        ├── enums.go                         # [NEW] Typed string enums
        ├── base.go                          # [NEW] Generic ListResponse, ErrorResponse, UpdatePayload
        ├── archer.go                        # [NEW] Archer domain models (Create, Read, Set, Filter, Update, ID)
        ├── auth.go                          # [NEW] Auth models (Create, Read, Set, Filter, Update, GoogleOneTap, AuthAuthenticated, AuthNeedsRegistration, Logout)
        ├── session.go                       # [NEW] Session models (Create, Read, Set, Filter, Update, ID)
        ├── target.go                        # [NEW] Target models (Create, Read, Set, Filter, Update)
        ├── slot.go                          # [NEW] Slot models (Create, Read, Set, Filter, Update, JoinRequest, JoinResponse, ReJoinRequest, LeaveRequest, ID, FullSlotInfo)
        ├── shot.go                          # [NEW] Shot models (Create, Read, Set, Filter, Update, ID)
        ├── face.go                          # [NEW] Face models (Spot, Ring, FaceMinimal, Face)
        ├── live_stats.go                    # [NEW] Live stats and WebSocket models (ShotScore, Stats, LiveStat, WebSocketMessage)
        ├── report.go                        # [NEW] Cross-domain analytics projections (SessionSummaryReport, ScoringTrend, ArcherPerformanceReport)
        └── model_test.go                    # [NEW] Unit tests verifying JSON serialization and round-tripping
```

---

### Task 1: Module Dependency, Shared Enums, and Base Types

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Create: `backend/internal/model/enums.go`
- Create: `backend/internal/model/base.go`
- Create: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: Standard library `encoding/json`
- Produces:
  - Types: `Gender`, `Bowstyle`, `SlotLetter`, `FaceType`, `AuthStatus`, `WSContentType`, `SessionStatus`
  - Generic types: `ListResponse[T any]`, `ErrorResponse`, `UpdatePayload[W, D any]`

- [ ] **Step 1: Write failing tests for enums and base types**

Create `backend/internal/model/model_test.go` with tests verifying JSON marshaling for enums and base models:

```go
package model_test

import (
	"encoding/json"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestEnums_JSON(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		expected string
	}{
		{"GenderMale", model.GenderMale, `"male"`},
		{"GenderFemale", model.GenderFemale, `"female"`},
		{"GenderNonBinary", model.GenderNonBinary, `"non_binary"`},
		{"GenderOther", model.GenderOther, `"other"`},
		{"GenderUnspecified", model.GenderUnspecified, `"unspecified"`},
		{"BowstyleRecurve", model.BowstyleRecurve, `"recurve"`},
		{"BowstyleCompound", model.BowstyleCompound, `"compound"`},
		{"BowstyleBarebow", model.BowstyleBarebow, `"barebow"`},
		{"BowstyleLongbow", model.BowstyleLongbow, `"longbow"`},
		{"SlotLetterA", model.SlotLetterA, `"A"`},
		{"SlotLetterB", model.SlotLetterB, `"B"`},
		{"FaceTypeWA40Full", model.FaceTypeWA40Full, `"wa_40cm_full"`},
		{"FaceTypeWA60TripleTriangular", model.FaceTypeWA60TripleTriangular, `"wa_60cm_triple_triangular"`},
		{"AuthStatusAuthenticated", model.AuthStatusAuthenticated, `"authenticated"`},
		{"AuthStatusNeedsRegistration", model.AuthStatusNeedsRegistration, `"needs_registration"`},
		{"WSContentTypeShotCreated", model.WSContentTypeShotCreated, `"shot.created"`},
		{"SessionStatusOpen", model.SessionStatusOpen, `"open"`},
		{"SessionStatusClosed", model.SessionStatusClosed, `"closed"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.val)
			if err != nil {
				t.Fatalf("failed to marshal %s: %v", tc.name, err)
			}
			if string(b) != tc.expected {
				t.Errorf("got %s, expected %s", string(b), tc.expected)
			}
		})
	}
}

func TestBaseTypes_JSON(t *testing.T) {
	t.Run("ListResponse", func(t *testing.T) {
		list := model.ListResponse[string]{
			Data:  []string{"foo", "bar"},
			Total: 2,
			Limit: 10,
			Page:  1,
		}
		b, err := json.Marshal(list)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		var decoded model.ListResponse[string]
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if decoded.Total != 2 || len(decoded.Data) != 2 || decoded.Data[0] != "foo" {
			t.Errorf("unexpected decoded list response: %+v", decoded)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		errResp := model.ErrorResponse{
			Detail: "Not found",
			Code:   "NOT_FOUND",
		}
		b, err := json.Marshal(errResp)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		var decoded model.ErrorResponse
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if decoded.Detail != "Not found" || decoded.Code != "NOT_FOUND" {
			t.Errorf("unexpected decoded error response: %+v", decoded)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: FAIL with `package github.com/jpmolinamatute/arch-stats/backend/internal/model is not in std` or `undefined: model.GenderMale`

- [ ] **Step 3: Implement `enums.go` and `base.go`**

Add `github.com/google/uuid` module dependency:
```bash
cd backend && go get github.com/google/uuid
```

Create `backend/internal/model/enums.go`:
```go
package model

// Gender represents the archer's gender identity.
type Gender string

const (
	GenderMale        Gender = "male"
	GenderFemale      Gender = "female"
	GenderNonBinary   Gender = "non_binary"
	GenderOther       Gender = "other"
	GenderUnspecified Gender = "unspecified"
)

// Bowstyle represents the equipment discipline.
type Bowstyle string

const (
	BowstyleRecurve  Bowstyle = "recurve"
	BowstyleCompound Bowstyle = "compound"
	BowstyleBarebow  Bowstyle = "barebow"
	BowstyleLongbow  Bowstyle = "longbow"
)

// SlotLetter represents the target lane division (A through D).
type SlotLetter string

const (
	SlotLetterA SlotLetter = "A"
	SlotLetterB SlotLetter = "B"
	SlotLetterC SlotLetter = "C"
	SlotLetterD SlotLetter = "D"
)

// FaceType represents World Archery approved target face dimensions and layouts.
type FaceType string

const (
	FaceTypeWA40Full               FaceType = "wa_40cm_full"
	FaceTypeWA60Full               FaceType = "wa_60cm_full"
	FaceTypeWA80Full               FaceType = "wa_80cm_full"
	FaceTypeWA122Full              FaceType = "wa_122cm_full"
	FaceTypeWA406Rings             FaceType = "wa_40cm_6rings"
	FaceTypeWA606Rings             FaceType = "wa_60cm_6rings"
	FaceTypeWA806Rings             FaceType = "wa_80cm_6rings"
	FaceTypeWA1226Rings            FaceType = "wa_122cm_6rings"
	FaceTypeWA40TripleVertical     FaceType = "wa_40cm_triple_vertical"
	FaceTypeWA60TripleTriangular   FaceType = "wa_60cm_triple_triangular"
	FaceTypeNone                   FaceType = "none"
)

// AuthStatus represents the outcome of a federated identity authentication attempt.
type AuthStatus string

const (
	AuthStatusAuthenticated     AuthStatus = "authenticated"
	AuthStatusNeedsRegistration AuthStatus = "needs_registration"
)

// WSContentType represents WebSocket event types broadcast to clients.
type WSContentType string

const (
	WSContentTypeShotCreated  WSContentType = "shot.created"
	WSContentTypeShotDeleted  WSContentType = "shot.deleted"
	WSContentTypeArrowCreated WSContentType = "arrow.created"
	WSContentTypeArrowDeleted WSContentType = "arrow.deleted"
)

// SessionStatus represents whether a shooting session is currently open or closed.
type SessionStatus string

const (
	SessionStatusOpen   SessionStatus = "open"
	SessionStatusClosed SessionStatus = "closed"
)
```

Create `backend/internal/model/base.go`:
```go
package model

// ListResponse wraps paginated query results.
type ListResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
	Limit int `json:"limit,omitempty"`
	Page  int `json:"page,omitempty"`
}

// ErrorResponse represents a standardized API error payload.
type ErrorResponse struct {
	Detail string `json:"detail"`
	Code   string `json:"code,omitempty"`
}

// UpdatePayload wraps where-clause criteria and updated data fields.
type UpdatePayload[W, D any] struct {
	Where W `json:"where"`
	Data  D `json:"data"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: PASS for `TestEnums_JSON` and `TestBaseTypes_JSON`.

- [ ] **Step 5: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/model/enums.go backend/internal/model/base.go backend/internal/model/model_test.go
git commit -m "feat(model): define enums and base response wrappers"
```

---

### Task 2: Archer and Auth Domain Models

**Files:**
- Create: `backend/internal/model/archer.go`
- Create: `backend/internal/model/auth.go`
- Modify: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: `model.Gender`, `model.Bowstyle`, `model.AuthStatus`, `github.com/google/uuid`, `time.Time`
- Produces:
  - Archer structs: `ArcherCreate`, `ArcherSet`, `ArcherFilter`, `ArcherUpdate`, `ArcherRead`, `ArcherID`
  - Auth structs: `AuthCreate`, `AuthFilter`, `AuthSet`, `AuthUpdate`, `AuthRead`, `GoogleOneTapRequest`, `AuthAuthenticated`, `AuthNeedsRegistration`, `LogoutResponse`, `AuthRegistrationRequest`

- [ ] **Step 1: Write failing tests for Archer and Auth models**

Append to `backend/internal/model/model_test.go`:

```go
func TestArcherModels_JSON(t *testing.T) {
	clubID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	picURL := "https://example.com/avatar.jpg"
	now := time.Now().UTC().Truncate(time.Second)

	create := model.ArcherCreate{
		FirstName:        "Robin",
		LastName:         "Hood",
		Email:            "robin@sherwood.org",
		DateOfBirth:      "1995-06-15",
		Gender:           model.GenderMale,
		Bowstyle:         model.BowstyleLongbow,
		DrawWeight:       45.5,
		ClubID:           &clubID,
		GooglePictureURL: &picURL,
		GoogleSubject:    "google-sub-12345",
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
		"bowstyle", "draw_weight", "club_id", "google_picture_url", "google_subject",
	}
	for _, k := range expectedKeys {
		if _, ok := createMap[k]; !ok {
			t.Errorf("expected JSON key %q missing in ArcherCreate serialization", k)
		}
	}

	archerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	read := model.ArcherRead{
		ArcherID:         archerID,
		FirstName:        create.FirstName,
		LastName:         create.LastName,
		Email:            create.Email,
		DateOfBirth:      create.DateOfBirth,
		Gender:           create.Gender,
		Bowstyle:         create.Bowstyle,
		DrawWeight:       create.DrawWeight,
		ClubID:           create.ClubID,
		GooglePictureURL: create.GooglePictureURL,
		GoogleSubject:    create.GoogleSubject,
		LastLoginAt:      now,
		CreatedAt:        now,
	}

	rb, err := json.Marshal(read)
	if err != nil {
		t.Fatalf("failed to marshal ArcherRead: %v", err)
	}

	var decodedRead model.ArcherRead
	if err := json.Unmarshal(rb, &decodedRead); err != nil {
		t.Fatalf("failed to unmarshal ArcherRead: %v", err)
	}

	if decodedRead.ArcherID != archerID || decodedRead.Email != create.Email || decodedRead.Gender != model.GenderMale {
		t.Errorf("mismatch in decoded ArcherRead: %+v", decodedRead)
	}
}

func TestAuthModels_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	archerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	authRead := model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: "jwt-token-xyz",
		ExpiresAt:   now.Add(time.Hour * 24),
		Archer: model.ArcherRead{
			ArcherID:      archerID,
			FirstName:     "Robin",
			LastName:      "Hood",
			Email:         "robin@sherwood.org",
			DateOfBirth:   "1995-06-15",
			Gender:        model.GenderMale,
			Bowstyle:      model.BowstyleLongbow,
			DrawWeight:    45.5,
			GoogleSubject: "google-sub-12345",
			LastLoginAt:   now,
			CreatedAt:     now,
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

	if decoded.Status != model.AuthStatusAuthenticated || decoded.AccessToken != "jwt-token-xyz" {
		t.Errorf("mismatch in decoded AuthAuthenticated: %+v", decoded)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: FAIL with `undefined: model.ArcherCreate`, `undefined: model.ArcherRead`, `undefined: model.AuthAuthenticated`

- [ ] **Step 3: Implement `archer.go` and `auth.go`**

Create `backend/internal/model/archer.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// ArcherCreate represents the payload required to create a new archer profile.
type ArcherCreate struct {
	FirstName        string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName         string     `json:"last_name" validate:"required,min=1,max=100"`
	Email            string     `json:"email" validate:"required,email"`
	DateOfBirth      string     `json:"date_of_birth" validate:"required"`
	Gender           Gender     `json:"gender" validate:"required"`
	Bowstyle         Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight       float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID           *uuid.UUID `json:"club_id,omitempty"`
	GooglePictureURL *string    `json:"google_picture_url,omitempty"`
	GoogleSubject    string     `json:"google_subject" validate:"required"`
}

// ArcherSet represents mutable fields when updating an archer profile.
type ArcherSet struct {
	FirstName        *string    `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName         *string    `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Gender           *Gender    `json:"gender,omitempty"`
	Bowstyle         *Bowstyle  `json:"bowstyle,omitempty"`
	DrawWeight       *float64   `json:"draw_weight,omitempty" validate:"omitempty,gt=0,lte=200"`
	ClubID           *uuid.UUID `json:"club_id,omitempty"`
	GooglePictureURL *string    `json:"google_picture_url,omitempty"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
}

// ArcherFilter represents criteria to query or select archers.
type ArcherFilter struct {
	ArcherID      *uuid.UUID `json:"archer_id,omitempty"`
	FirstName     *string    `json:"first_name,omitempty"`
	LastName      *string    `json:"last_name,omitempty"`
	Gender        *Gender    `json:"gender,omitempty"`
	Bowstyle      *Bowstyle  `json:"bowstyle,omitempty"`
	DrawWeight    *float64   `json:"draw_weight,omitempty"`
	ClubID        *uuid.UUID `json:"club_id,omitempty"`
	GoogleSubject *string    `json:"google_subject,omitempty"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

// ArcherUpdate wraps target filter criteria and field updates for archers.
type ArcherUpdate struct {
	Where ArcherFilter `json:"where"`
	Data  ArcherSet    `json:"data"`
}

// ArcherRead represents the full persisted archer domain model.
type ArcherRead struct {
	ArcherID         uuid.UUID  `json:"archer_id"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Email            string     `json:"email"`
	DateOfBirth      string     `json:"date_of_birth"`
	Gender           Gender     `json:"gender"`
	Bowstyle         Bowstyle   `json:"bowstyle"`
	DrawWeight       float64    `json:"draw_weight"`
	ClubID           *uuid.UUID `json:"club_id,omitempty"`
	GooglePictureURL *string    `json:"google_picture_url,omitempty"`
	GoogleSubject    string     `json:"google_subject"`
	LastLoginAt      time.Time  `json:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// ArcherID represents a standalone archer identifier wrapper.
type ArcherID struct {
	ArcherID uuid.UUID `json:"archer_id"`
}
```

Create `backend/internal/model/auth.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// AuthCreate represents the creation of a server-side authentication session.
type AuthCreate struct {
	ArcherID         uuid.UUID `json:"archer_id"`
	SessionTokenHash []byte    `json:"session_token_hash"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	UA               *string   `json:"ua,omitempty"`
	IPInet           *string   `json:"ip_inet,omitempty"`
}

// AuthFilter represents criteria to query authentication sessions.
type AuthFilter struct {
	SessionTokenHash []byte     `json:"session_token_hash,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}

// AuthSet represents updates to an authentication session (e.g. revocation).
type AuthSet struct {
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// AuthUpdate wraps filter and update data for auth sessions.
type AuthUpdate struct {
	Where AuthFilter `json:"where"`
	Data  AuthSet    `json:"data"`
}

// AuthRead represents a persisted authentication session record.
type AuthRead struct {
	AuthID           uuid.UUID  `json:"auth_id"`
	ArcherID         uuid.UUID  `json:"archer_id"`
	SessionTokenHash []byte     `json:"session_token_hash"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	UA               *string    `json:"ua,omitempty"`
	IPInet           *string    `json:"ip_inet,omitempty"`
}

// GoogleOneTapRequest represents the incoming ID token from Google One Tap.
type GoogleOneTapRequest struct {
	Credential string `json:"credential" validate:"required,min=10"`
}

// AuthAuthenticated represents a successful authentication response.
type AuthAuthenticated struct {
	Status      AuthStatus `json:"status"`
	AccessToken string     `json:"access_token"`
	ExpiresAt   time.Time  `json:"expires_at"`
	Archer      ArcherRead `json:"archer"`
}

// AuthNeedsRegistration represents Google identity verified but requiring archer registration.
type AuthNeedsRegistration struct {
	Status              AuthStatus `json:"status"`
	GoogleEmail         string     `json:"google_email"`
	GoogleSubject       string     `json:"google_subject"`
	GivenName           *string    `json:"given_name,omitempty"`
	FamilyName          *string    `json:"family_name,omitempty"`
	GivenNameProvided   bool       `json:"given_name_provided"`
	FamilyNameProvided  bool       `json:"family_name_provided"`
	PictureURL          *string    `json:"picture_url,omitempty"`
}

// LogoutResponse indicates logout operation success.
type LogoutResponse struct {
	Success bool `json:"success"`
}

// AuthRegistrationRequest represents profile information submitted on first registration.
type AuthRegistrationRequest struct {
	Credential  string     `json:"credential" validate:"required,min=10"`
	DateOfBirth string     `json:"date_of_birth" validate:"required"`
	Gender      Gender     `json:"gender" validate:"required"`
	Bowstyle    Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight  float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID      *uuid.UUID `json:"club_id,omitempty"`
	FirstName   *string    `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName    *string    `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: PASS for all tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/model/archer.go backend/internal/model/auth.go backend/internal/model/model_test.go
git commit -m "feat(model): define archer and auth domain model structs"
```

---

### Task 3: Session, Target, and Slot Domain Models

**Files:**
- Create: `backend/internal/model/session.go`
- Create: `backend/internal/model/target.go`
- Create: `backend/internal/model/slot.go`
- Modify: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: `model.FaceType`, `model.SlotLetter`, `model.Bowstyle`, `github.com/google/uuid`, `time.Time`
- Produces:
  - Session structs: `SessionCreate`, `SessionSet`, `SessionFilter`, `SessionUpdate`, `SessionRead`, `SessionID`
  - Target structs: `TargetCreate`, `TargetSet`, `TargetFilter`, `TargetUpdate`, `TargetRead`
  - Slot structs: `SlotCreate`, `SlotSet`, `SlotFilter`, `SlotUpdate`, `SlotRead`, `SlotJoinRequest`, `SlotJoinResponse`, `SlotReJoinRequest`, `SlotLeaveRequest`, `SlotID`, `FullSlotInfo`

- [ ] **Step 1: Write failing tests for Session, Target, and Slot models**

Append to `backend/internal/model/model_test.go`:

```go
func TestSessionModels_JSON(t *testing.T) {
	ownerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	now := time.Now().UTC().Truncate(time.Second)

	create := model.SessionCreate{
		OwnerArcherID:   ownerID,
		SessionLocation: "Sherwood Forest Range",
		IsIndoor:        false,
		IsOpened:        true,
	}

	b, err := json.Marshal(create)
	if err != nil {
		t.Fatalf("failed to marshal SessionCreate: %v", err)
	}

	var createMap map[string]any
	if err := json.Unmarshal(b, &createMap); err != nil {
		t.Fatalf("failed to unmarshal SessionCreate JSON: %v", err)
	}

	for _, k := range []string{"owner_archer_id", "session_location", "is_indoor", "is_opened"} {
		if _, ok := createMap[k]; !ok {
			t.Errorf("expected key %q missing in SessionCreate", k)
		}
	}

	read := model.SessionRead{
		SessionID:       sessionID,
		OwnerArcherID:   ownerID,
		SessionLocation: create.SessionLocation,
		IsIndoor:        create.IsIndoor,
		IsOpened:        create.IsOpened,
		CreatedAt:       now,
		ClosedAt:        nil,
	}

	rb, err := json.Marshal(read)
	if err != nil {
		t.Fatalf("failed to marshal SessionRead: %v", err)
	}

	var decodedRead model.SessionRead
	if err := json.Unmarshal(rb, &decodedRead); err != nil {
		t.Fatalf("failed to unmarshal SessionRead: %v", err)
	}

	if decodedRead.SessionID != sessionID || decodedRead.ClosedAt != nil {
		t.Errorf("mismatch in decoded SessionRead: %+v", decodedRead)
	}
}

func TestTargetAndSlotModels_JSON(t *testing.T) {
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	targetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	slotID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	archerID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	now := time.Now().UTC().Truncate(time.Second)

	fullSlot := model.FullSlotInfo{
		SlotID:          slotID,
		TargetID:        targetID,
		ArcherID:        archerID,
		SessionID:       sessionID,
		SlotLetter:      model.SlotLetterA,
		Lane:            1,
		Distance:        18,
		Slot:            "1A",
		FaceType:        model.FaceTypeWA40Full,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      38.0,
		IsShooting:      true,
		IntervalSeconds: 20,
		CreatedAt:       now,
	}

	b, err := json.Marshal(fullSlot)
	if err != nil {
		t.Fatalf("failed to marshal FullSlotInfo: %v", err)
	}

	var decoded model.FullSlotInfo
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FullSlotInfo: %v", err)
	}

	if decoded.Slot != "1A" || decoded.Lane != 1 || decoded.SlotLetter != model.SlotLetterA {
		t.Errorf("mismatch in decoded FullSlotInfo: %+v", decoded)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: FAIL with `undefined: model.SessionCreate`, `undefined: model.FullSlotInfo`

- [ ] **Step 3: Implement `session.go`, `target.go`, and `slot.go`**

Create `backend/internal/model/session.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// SessionCreate represents the payload to create a new shooting session.
type SessionCreate struct {
	OwnerArcherID   uuid.UUID `json:"owner_archer_id" validate:"required"`
	SessionLocation string    `json:"session_location" validate:"required,min=1,max=255"`
	IsIndoor        bool      `json:"is_indoor"`
	IsOpened        bool      `json:"is_opened"`
}

// SessionSet represents mutable fields when updating a session.
type SessionSet struct {
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	SessionLocation *string    `json:"session_location,omitempty" validate:"omitempty,min=1,max=255"`
	IsOpened        *bool      `json:"is_opened,omitempty"`
	IsIndoor        *bool      `json:"is_indoor,omitempty"`
}

// SessionFilter represents criteria to query shooting sessions.
type SessionFilter struct {
	SessionID       *uuid.UUID `json:"session_id,omitempty"`
	OwnerArcherID   *uuid.UUID `json:"owner_archer_id,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	SessionLocation *string    `json:"session_location,omitempty"`
	IsOpened        *bool      `json:"is_opened,omitempty"`
	IsIndoor        *bool      `json:"is_indoor,omitempty"`
}

// SessionUpdate wraps filter criteria and mutation data for sessions.
type SessionUpdate struct {
	Where SessionFilter `json:"where"`
	Data  SessionSet    `json:"data"`
}

// SessionRead represents a persisted shooting session.
type SessionRead struct {
	SessionID       uuid.UUID  `json:"session_id"`
	OwnerArcherID   uuid.UUID  `json:"owner_archer_id"`
	SessionLocation string     `json:"session_location"`
	IsIndoor        bool       `json:"is_indoor"`
	IsOpened        bool       `json:"is_opened"`
	CreatedAt       time.Time  `json:"created_at"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
}

// SessionID represents a standalone session identifier wrapper.
type SessionID struct {
	SessionID *uuid.UUID `json:"session_id,omitempty"`
}
```

Create `backend/internal/model/target.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// TargetCreate represents the payload to create a lane target in a session.
type TargetCreate struct {
	SessionID uuid.UUID `json:"session_id" validate:"required"`
	Distance  int       `json:"distance" validate:"required,gte=1,lte=100"`
	Lane      int       `json:"lane" validate:"required,gte=1,lte=100"`
}

// TargetSet represents updates to a target's distance or lane.
type TargetSet struct {
	Distance *int `json:"distance,omitempty" validate:"omitempty,gte=1,lte=100"`
	Lane     *int `json:"lane,omitempty" validate:"omitempty,gte=1,lte=100"`
}

// TargetFilter represents criteria to filter targets.
type TargetFilter struct {
	TargetID  *uuid.UUID `json:"target_id,omitempty"`
	SessionID *uuid.UUID `json:"session_id,omitempty"`
	Distance  *int       `json:"distance,omitempty"`
	Lane      *int       `json:"lane,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// TargetUpdate wraps filter criteria and mutation data for targets.
type TargetUpdate struct {
	Where TargetFilter `json:"where"`
	Data  TargetSet    `json:"data"`
}

// TargetRead represents a persisted target with lane, distance, and optional occupancy count.
type TargetRead struct {
	TargetID  uuid.UUID `json:"target_id"`
	SessionID uuid.UUID `json:"session_id"`
	Distance  int       `json:"distance"`
	Lane      int       `json:"lane"`
	Occupied  *int      `json:"occupied,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
```

Create `backend/internal/model/slot.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// SlotCreate represents the payload to directly assign an archer to a target slot.
type SlotCreate struct {
	ArcherID        uuid.UUID  `json:"archer_id" validate:"required"`
	SessionID       uuid.UUID  `json:"session_id" validate:"required"`
	TargetID        uuid.UUID  `json:"target_id" validate:"required"`
	SlotLetter      SlotLetter `json:"slot_letter" validate:"required"`
	FaceType        FaceType   `json:"face_type" validate:"required"`
	Bowstyle        Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight      float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty" validate:"omitempty,gte=3,lte=10"`
	IntervalSeconds int        `json:"interval_seconds" validate:"gte=1,lte=100"`
}

// SlotJoinRequest represents a request by an archer to join a session at a preferred distance.
type SlotJoinRequest struct {
	ArcherID        uuid.UUID  `json:"archer_id" validate:"required"`
	SessionID       uuid.UUID  `json:"session_id" validate:"required"`
	Distance        int        `json:"distance" validate:"required,gte=1,lte=100"`
	FaceType        FaceType   `json:"face_type" validate:"required"`
	Bowstyle        Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight      float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty" validate:"omitempty,gte=3,lte=10"`
	IntervalSeconds int        `json:"interval_seconds" validate:"gte=1,lte=100"`
}

// SlotJoinResponse represents the assigned slot code and slot ID returned upon joining.
type SlotJoinResponse struct {
	SlotID uuid.UUID `json:"slot_id"`
	Slot   string    `json:"slot"`
}

// SlotReJoinRequest represents a request to re-enter an existing slot assignment.
type SlotReJoinRequest struct {
	SlotID    uuid.UUID `json:"slot_id" validate:"required"`
	SessionID uuid.UUID `json:"session_id" validate:"required"`
	ArcherID  uuid.UUID `json:"archer_id" validate:"required"`
}

// SlotLeaveRequest represents an archer leaving their slot.
type SlotLeaveRequest struct {
	ArcherID  uuid.UUID `json:"archer_id" validate:"required"`
	SessionID uuid.UUID `json:"session_id" validate:"required"`
}

// SlotID represents a standalone slot identifier wrapper.
type SlotID struct {
	SlotID uuid.UUID `json:"slot_id"`
}

// SlotRead represents a persisted slot assignment record.
type SlotRead struct {
	SlotID          uuid.UUID  `json:"slot_id"`
	TargetID        uuid.UUID  `json:"target_id"`
	ArcherID        uuid.UUID  `json:"archer_id"`
	SessionID       uuid.UUID  `json:"session_id"`
	SlotLetter      SlotLetter `json:"slot_letter"`
	Slot            *string    `json:"slot,omitempty"`
	FaceType        FaceType   `json:"face_type"`
	Bowstyle        Bowstyle   `json:"bowstyle"`
	DrawWeight      float64    `json:"draw_weight"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty"`
	IntervalSeconds int        `json:"interval_seconds"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

// SlotSet represents mutable fields on an existing slot assignment.
type SlotSet struct {
	IsShooting      *bool       `json:"is_shooting,omitempty"`
	FaceType        *FaceType   `json:"face_type,omitempty"`
	SlotLetter      *SlotLetter `json:"slot_letter,omitempty"`
	ShotPerRound    *int        `json:"shot_per_round,omitempty" validate:"omitempty,gte=3,lte=10"`
	IntervalSeconds *int        `json:"interval_seconds,omitempty" validate:"omitempty,gte=1,lte=100"`
}

// SlotFilter represents criteria to query slot assignments.
type SlotFilter struct {
	SlotID       *uuid.UUID  `json:"slot_id,omitempty"`
	TargetID     *uuid.UUID  `json:"target_id,omitempty"`
	ArcherID     *uuid.UUID  `json:"archer_id,omitempty"`
	SessionID    *uuid.UUID  `json:"session_id,omitempty"`
	SlotLetter   *SlotLetter `json:"slot_letter,omitempty"`
	IsShooting   *bool       `json:"is_shooting,omitempty"`
	ShotPerRound *int        `json:"shot_per_round,omitempty"`
	CreatedAt    *time.Time  `json:"created_at,omitempty"`
}

// SlotUpdate wraps filter criteria and mutation data for slot assignments.
type SlotUpdate struct {
	Where SlotFilter `json:"where"`
	Data  SlotSet    `json:"data"`
}

// FullSlotInfo represents a comprehensive denormalized slot assignment including lane, distance, and composite code.
type FullSlotInfo struct {
	SlotID          uuid.UUID  `json:"slot_id"`
	TargetID        uuid.UUID  `json:"target_id"`
	ArcherID        uuid.UUID  `json:"archer_id"`
	SessionID       uuid.UUID  `json:"session_id"`
	SlotLetter      SlotLetter `json:"slot_letter"`
	Lane            int        `json:"lane"`
	Distance        int        `json:"distance"`
	Slot            string     `json:"slot"`
	FaceType        FaceType   `json:"face_type"`
	Bowstyle        Bowstyle   `json:"bowstyle"`
	DrawWeight      float64    `json:"draw_weight"`
	ClubID          *uuid.UUID `json:"club_id,omitempty"`
	IsShooting      bool       `json:"is_shooting"`
	ShotPerRound    *int       `json:"shot_per_round,omitempty"`
	IntervalSeconds int        `json:"interval_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: PASS for all tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/model/session.go backend/internal/model/target.go backend/internal/model/slot.go backend/internal/model/model_test.go
git commit -m "feat(model): define session, target, and slot domain models"
```

---

### Task 4: Shot, Face, Live Stats, and Reporting Models

**Files:**
- Create: `backend/internal/model/shot.go`
- Create: `backend/internal/model/face.go`
- Create: `backend/internal/model/live_stats.go`
- Create: `backend/internal/model/report.go`
- Modify: `backend/internal/model/model_test.go`

**Interfaces:**
- Consumes: `model.FaceType`, `model.WSContentType`, `github.com/google/uuid`, `time.Time`
- Produces:
  - Shot structs: `ShotCreate`, `ShotSet`, `ShotFilter`, `ShotUpdate`, `ShotRead`, `ShotID`
  - Face structs: `Spot`, `Ring`, `FaceMinimal`, `Face`
  - Live stats structs: `ShotScore`, `Stats`, `LiveStat`, `WebSocketMessage`
  - Report structs: `SessionSummaryReport`, `ScoringTrend`, `ArcherPerformanceReport`

- [ ] **Step 1: Write failing tests for Shot, Face, Live Stats, and Report models**

Append to `backend/internal/model/model_test.go`:

```go
func TestShotAndFaceModels_JSON(t *testing.T) {
	shotID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	slotID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	arrowID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Now().UTC().Truncate(time.Second)

	x := 12.5
	y := -8.3
	score := 10
	shot := model.ShotRead{
		ShotID:    shotID,
		SlotID:    slotID,
		X:         &x,
		Y:         &y,
		IsX:       true,
		Score:     &score,
		ArrowID:   &arrowID,
		CreatedAt: now,
	}

	b, err := json.Marshal(shot)
	if err != nil {
		t.Fatalf("failed to marshal ShotRead: %v", err)
	}

	var decodedShot model.ShotRead
	if err := json.Unmarshal(b, &decodedShot); err != nil {
		t.Fatalf("failed to unmarshal ShotRead: %v", err)
	}

	if decodedShot.ShotID != shotID || *decodedShot.X != 12.5 || !decodedShot.IsX || *decodedShot.Score != 10 {
		t.Errorf("mismatch in decoded ShotRead: %+v", decodedShot)
	}

	face := model.Face{
		FaceType: model.FaceTypeWA40Full,
		FaceName: "WA 40cm Full Face",
		Spots: []model.Spot{
			{XOffset: 0.0, YOffset: 0.0, Diameter: 400.0},
		},
		Rings: []model.Ring{
			{DataScore: 10, Fill: "#FFD700", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
		ViewBox:     400.0,
		RenderCross: true,
	}

	fb, err := json.Marshal(face)
	if err != nil {
		t.Fatalf("failed to marshal Face: %v", err)
	}

	var decodedFace model.Face
	if err := json.Unmarshal(fb, &decodedFace); err != nil {
		t.Fatalf("failed to unmarshal Face: %v", err)
	}

	if decodedFace.FaceType != model.FaceTypeWA40Full || len(decodedFace.Spots) != 1 || len(decodedFace.Rings) != 1 {
		t.Errorf("mismatch in decoded Face: %+v", decodedFace)
	}
}

func TestLiveStatsAndReports_JSON(t *testing.T) {
	slotID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shotID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	archerID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	now := time.Now().UTC().Truncate(time.Second)

	wsMsg := model.WebSocketMessage{
		TS:          now,
		ContentType: model.WSContentTypeShotCreated,
		Content: model.LiveStat{
			Scores: []model.ShotScore{
				{ShotID: shotID, Score: 10, IsX: true, CreatedAt: now},
			},
			Stats: model.Stats{
				SlotID:        slotID,
				NumberOfShots: 1,
				TotalScore:    10,
				MaxScore:      10,
				Mean:          10.0,
			},
		},
	}

	b, err := json.Marshal(wsMsg)
	if err != nil {
		t.Fatalf("failed to marshal WebSocketMessage: %v", err)
	}

	var decodedMsg model.WebSocketMessage
	if err := json.Unmarshal(b, &decodedMsg); err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	if decodedMsg.ContentType != model.WSContentTypeShotCreated || decodedMsg.Content.Stats.TotalScore != 10 {
		t.Errorf("mismatch in decoded WebSocketMessage: %+v", decodedMsg)
	}

	report := model.SessionSummaryReport{
		SessionID:       sessionID,
		SessionLocation: "Sherwood Forest",
		TotalShots:      60,
		AverageScore:    9.45,
		StartedAt:       now,
	}

	rb, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal SessionSummaryReport: %v", err)
	}

	var decodedReport model.SessionSummaryReport
	if err := json.Unmarshal(rb, &decodedReport); err != nil {
		t.Fatalf("failed to unmarshal SessionSummaryReport: %v", err)
	}

	if decodedReport.SessionID != sessionID || decodedReport.TotalShots != 60 {
		t.Errorf("mismatch in decoded SessionSummaryReport: %+v", decodedReport)
	}

	archPerf := model.ArcherPerformanceReport{
		ArcherID:        archerID,
		TotalSessions:   12,
		TotalShots:      720,
		AverageScore:    9.2,
		BestScore:       300,
		TotalXCount:     154,
	}

	ab, err := json.Marshal(archPerf)
	if err != nil {
		t.Fatalf("failed to marshal ArcherPerformanceReport: %v", err)
	}

	var decodedPerf model.ArcherPerformanceReport
	if err := json.Unmarshal(ab, &decodedPerf); err != nil {
		t.Fatalf("failed to unmarshal ArcherPerformanceReport: %v", err)
	}

	if decodedPerf.ArcherID != archerID || decodedPerf.TotalXCount != 154 {
		t.Errorf("mismatch in decoded ArcherPerformanceReport: %+v", decodedPerf)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: FAIL with `undefined: model.ShotRead`, `undefined: model.Face`, `undefined: model.WebSocketMessage`, `undefined: model.SessionSummaryReport`

- [ ] **Step 3: Implement `shot.go`, `face.go`, `live_stats.go`, and `report.go`**

Create `backend/internal/model/shot.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// ShotCreate represents the payload to register a single shot.
// Coordinates x, y, and score must either all be present or all be nil.
type ShotCreate struct {
	SlotID    uuid.UUID  `json:"slot_id" validate:"required"`
	X         *float64   `json:"x,omitempty"`
	Y         *float64   `json:"y,omitempty"`
	IsX       bool       `json:"is_x"`
	Score     *int       `json:"score,omitempty" validate:"omitempty,gte=0,lte=10"`
	ArrowID   *uuid.UUID `json:"arrow_id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// ShotSet represents updates to shot fields (placeholder; shots are immutable).
type ShotSet struct{}

// ShotFilter represents criteria to query shots.
type ShotFilter struct {
	ShotID    *uuid.UUID `json:"shot_id,omitempty"`
	SlotID    *uuid.UUID `json:"slot_id,omitempty"`
	X         *float64   `json:"x,omitempty"`
	Y         *float64   `json:"y,omitempty"`
	Score     *int       `json:"score,omitempty" validate:"omitempty,gte=0,lte=10"`
	ArrowID   *uuid.UUID `json:"arrow_id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// ShotUpdate wraps filter and update data for shots.
type ShotUpdate struct {
	Where ShotFilter `json:"where"`
	Data  ShotSet    `json:"data"`
}

// ShotRead represents a persisted shot record.
type ShotRead struct {
	ShotID    uuid.UUID  `json:"shot_id"`
	SlotID    uuid.UUID  `json:"slot_id"`
	X         *float64   `json:"x,omitempty"`
	Y         *float64   `json:"y,omitempty"`
	IsX       bool       `json:"is_x"`
	Score     *int       `json:"score,omitempty"`
	ArrowID   *uuid.UUID `json:"arrow_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ShotID represents a standalone shot identifier wrapper.
type ShotID struct {
	ShotID uuid.UUID `json:"shot_id"`
}
```

Create `backend/internal/model/face.go`:
```go
package model

// Spot represents the position and diameter of an individual target spot on a face.
type Spot struct {
	XOffset  float64 `json:"x_offset"`
	YOffset  float64 `json:"y_offset"`
	Diameter float64 `json:"diameter"`
}

// Ring represents a concentric scoring zone on a target face.
type Ring struct {
	DataScore   int     `json:"data_score"`
	Fill        string  `json:"fill"`
	R           float64 `json:"r"`
	Stroke      string  `json:"stroke"`
	StrokeWidth float64 `json:"stroke_width"`
}

// FaceMinimal provides basic identifying information about a target face.
type FaceMinimal struct {
	FaceType FaceType `json:"face_type"`
	FaceName string   `json:"face_name"`
}

// Face describes the full geometry, styling, and spots of a target face.
type Face struct {
	FaceType    FaceType `json:"face_type"`
	FaceName    string   `json:"face_name"`
	Spots       []Spot   `json:"spots"`
	Rings       []Ring   `json:"rings"`
	ViewBox     float64  `json:"viewBox"`
	RenderCross bool     `json:"render_cross"`
}
```

Create `backend/internal/model/live_stats.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// ShotScore represents score and timing information for an individual shot in live views.
type ShotScore struct {
	ShotID    uuid.UUID `json:"shot_id"`
	Score     int       `json:"score" validate:"gte=0,lte=10"`
	IsX       bool      `json:"is_x"`
	CreatedAt time.Time `json:"created_at"`
}

// Stats represents real-time aggregate scoring metrics for a single slot.
type Stats struct {
	SlotID        uuid.UUID `json:"slot_id"`
	NumberOfShots int       `json:"number_of_shots" validate:"gte=0"`
	TotalScore    int       `json:"total_score" validate:"gte=0"`
	MaxScore      int       `json:"max_score" validate:"gte=0"`
	Mean          float64   `json:"mean"`
}

// LiveStat aggregates recent shots and running statistics for a slot.
type LiveStat struct {
	Scores []ShotScore `json:"scores"`
	Stats  Stats       `json:"stats"`
}

// WebSocketMessage represents a live update broadcast over WebSockets.
type WebSocketMessage struct {
	TS          time.Time     `json:"ts"`
	ContentType WSContentType `json:"content_type"`
	Content     LiveStat      `json:"content"`
}
```

Create `backend/internal/model/report.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// SessionSummaryReport represents aggregated performance statistics for a single completed session.
type SessionSummaryReport struct {
	SessionID       uuid.UUID  `json:"session_id"`
	SessionLocation string     `json:"session_location"`
	TotalShots      int        `json:"total_shots"`
	AverageScore    float64    `json:"average_score"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
}

// ScoringTrend represents a discrete data point in a time-series of an archer's scoring progression.
type ScoringTrend struct {
	SessionID    uuid.UUID `json:"session_id"`
	Timestamp    time.Time `json:"timestamp"`
	AverageScore float64   `json:"average_score"`
	TotalShots   int       `json:"total_shots"`
}

// ArcherPerformanceReport represents historical career metrics across multiple sessions.
type ArcherPerformanceReport struct {
	ArcherID      uuid.UUID `json:"archer_id"`
	TotalSessions int       `json:"total_sessions"`
	TotalShots    int       `json:"total_shots"`
	AverageScore  float64   `json:"average_score"`
	BestScore     int       `json:"best_score"`
	TotalXCount   int       `json:"total_x_count"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/model/... -v
```
Expected: PASS for all tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/model/shot.go backend/internal/model/face.go backend/internal/model/live_stats.go backend/internal/model/report.go backend/internal/model/model_test.go
git commit -m "feat(model): define shot, face, live_stats, and report models"
```

---

### Task 5: Final Cleanup, Linting, Verification, and Checklist Update

**Files:**
- Delete: `backend/internal/model/.gitkeep`
- Modify: `docs/go_refactor/tasks/007-model_structs.md`

**Interfaces:**
- Consumes: All packages in `backend/internal/model`
- Produces: Verified Go domain model package with 100% clean build, tests, and linters

- [ ] **Step 1: Delete `.gitkeep`**

```bash
rm -f backend/internal/model/.gitkeep
```

- [ ] **Step 2: Run Go test suite with race detector**

Run:
```bash
cd backend && go test -race ./... -v
```
Expected: All tests pass with zero failures.

- [ ] **Step 3: Run `go vet` across the repository**

Run:
```bash
cd backend && go vet ./...
```
Expected: Zero issues reported.

- [ ] **Step 4: Run `golangci-lint`**

Run:
```bash
cd backend && golangci-lint run ./...
```
Expected: Zero linter issues reported.

- [ ] **Step 5: Verify build**

Run:
```bash
cd backend && go build ./...
```
Expected: Exit code 0, all packages compile cleanly.

- [ ] **Step 6: Update task status document**

Update `docs/go_refactor/tasks/007-model_structs.md` checking off all acceptance criteria and steps.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/model/.gitkeep docs/go_refactor/tasks/007-model_structs.md
git commit -m "chore(model): clean up gitkeep and mark task 007 complete"
```
