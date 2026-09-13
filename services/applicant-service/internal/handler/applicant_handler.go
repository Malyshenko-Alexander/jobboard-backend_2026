package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/study/jobboard/applicant-service/internal/authtoken"
	"github.com/study/jobboard/applicant-service/internal/model"
	"github.com/study/jobboard/applicant-service/internal/service"
)

type applicantService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (model.ApplicantProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.ApplicantProfile, error)
	GetResume(ctx context.Context, userID uuid.UUID) (model.Resume, error)
	UpsertResume(ctx context.Context, userID uuid.UUID, req model.UpsertResumeRequest) (model.Resume, error)
}

type userCreatedHandler interface {
	HandleUserCreated(ctx context.Context, raw []byte) error
}

// ApplicantHandler exposes HTTP endpoints for applicant cabinet.
type ApplicantHandler struct {
	svc      applicantService
	tokens   *authtoken.Manager
	consumer userCreatedHandler
}

func NewApplicantHandler(svc applicantService, tokens *authtoken.Manager, consumer userCreatedHandler) *ApplicantHandler {
	return &ApplicantHandler{svc: svc, tokens: tokens, consumer: consumer}
}

// GetProfile godoc
// @Summary      Get applicant profile
// @Tags         applicant
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.ApplicantProfile
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/applicant/profile [get]
func (h *ApplicantHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	profile, err := h.svc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary      Update applicant profile
// @Tags         applicant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.UpdateProfileRequest  true  "Profile payload"
// @Success      200   {object}  model.ApplicantProfile
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Router       /api/v1/applicant/profile [put]
func (h *ApplicantHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	claims := ClaimsFromContext(r.Context())
	profile, err := h.svc.UpdateProfile(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// GetResume godoc
// @Summary      Get active resume
// @Tags         applicant
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.Resume
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/applicant/resume [get]
func (h *ApplicantHandler) GetResume(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	resume, err := h.svc.GetResume(r.Context(), claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resume)
}

// UpsertResume godoc
// @Summary      Create or update active resume
// @Tags         applicant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.UpsertResumeRequest  true  "Resume payload"
// @Success      200   {object}  model.Resume
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Router       /api/v1/applicant/resume [put]
func (h *ApplicantHandler) UpsertResume(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	claims := ClaimsFromContext(r.Context())
	resume, err := h.svc.UpsertResume(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resume)
}

// UserCreatedHook godoc
// @Summary      Stub hook for user.created event
// @Description  Temporary HTTP stand-in for RabbitMQ consumer. Remove when real AMQP consumer is wired.
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        body  body      model.UserCreatedEvent  true  "user.created payload"
// @Success      204
// @Failure      400  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/internal/events/user-created [post]
func (h *ApplicantHandler) UserCreatedHook(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.consumer.HandleUserCreated(r.Context(), raw); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to handle event")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AuthMiddleware validates Bearer JWT and requires applicant role.
func (h *ApplicantHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := h.tokens.Parse(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		if claims.Role != authtoken.RoleApplicant {
			writeError(w, http.StatusForbidden, service.ErrForbiddenRole.Error())
			return
		}

		ctx := ContextWithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrForbiddenRole):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}
