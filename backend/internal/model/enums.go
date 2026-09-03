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
	FaceTypeWA40Full             FaceType = "wa_40cm_full"
	FaceTypeWA60Full             FaceType = "wa_60cm_full"
	FaceTypeWA80Full             FaceType = "wa_80cm_full"
	FaceTypeWA122Full            FaceType = "wa_122cm_full"
	FaceTypeWA406Rings           FaceType = "wa_40cm_6rings"
	FaceTypeWA606Rings           FaceType = "wa_60cm_6rings"
	FaceTypeWA806Rings           FaceType = "wa_80cm_6rings"
	FaceTypeWA1226Rings          FaceType = "wa_122cm_6rings"
	FaceTypeWA40TripleVertical   FaceType = "wa_40cm_triple_vertical"
	FaceTypeWA60TripleTriangular FaceType = "wa_60cm_triple_triangular"
	FaceTypeNone                 FaceType = "none"
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
