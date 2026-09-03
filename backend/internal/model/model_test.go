package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
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

