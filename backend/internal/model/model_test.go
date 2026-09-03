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

