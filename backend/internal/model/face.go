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

// FaceRead is an alias for Face to support repository and service layer naming conventions.
type FaceRead = Face
