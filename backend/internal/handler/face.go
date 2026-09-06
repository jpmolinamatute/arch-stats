package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// FaceService defines the domain operations required by FaceHandler.
type FaceService interface {
	ListAll(ctx context.Context) ([]model.FaceRead, error)
	GetByID(ctx context.Context, id string) (*model.FaceRead, error)
}

// FaceHandler manages HTTP endpoints for target face catalog definitions.
type FaceHandler struct {
	faceSvc FaceService
}

// NewFaceHandler constructs a FaceHandler with service dependency injection.
func NewFaceHandler(faceSvc FaceService) *FaceHandler {
	return &FaceHandler{
		faceSvc: faceSvc,
	}
}

// Routes registers all face catalog endpoints on the provided chi Router.
func (h *FaceHandler) Routes(r chi.Router) {
	r.Get("/", h.ListFaces)
	r.Get("/{face_type}", h.GetFace)
}

// ListFaces godoc
// @Summary     List Faces
// @Description Returns a list of target face summaries (face_type and face_name). Public endpoint.
// @Tags        Faces
// @Produce     json
// @Success     200 {array} model.FaceMinimal
// @Router      /faces [get]
func (h *FaceHandler) ListFaces(w http.ResponseWriter, r *http.Request) {
	faces, err := h.faceSvc.ListAll(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}

	summaries := make([]model.FaceMinimal, len(faces))
	for i, f := range faces {
		summaries[i] = model.FaceMinimal{
			FaceType: f.FaceType,
			FaceName: f.FaceName,
		}
	}

	_ = writeJSON(w, http.StatusOK, summaries)
}

// GetFace godoc
// @Summary     Get Face
// @Description Returns the full geometry and scoring layout for the requested face type. Public endpoint.
// @Tags        Faces
// @Produce     json
// @Param       face_type path string true "Face type identifier"
// @Success     200 {object} model.Face
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Router      /faces/{face_type} [get]
func (h *FaceHandler) GetFace(w http.ResponseWriter, r *http.Request) {
	faceType := getURLParam(r, "face_type")
	if faceType == "" {
		faceType = getURLParam(r, "id")
	}

	if strings.TrimSpace(faceType) == "" {
		writeAppError(w, apperror.Wrap(apperror.ErrValidation, "face_type is required"))
		return
	}

	face, err := h.faceSvc.GetByID(r.Context(), faceType)
	if err != nil {
		writeAppError(w, err)
		return
	}
	if face == nil {
		writeAppError(w, apperror.ErrNotFound)
		return
	}

	_ = writeJSON(w, http.StatusOK, face)
}
