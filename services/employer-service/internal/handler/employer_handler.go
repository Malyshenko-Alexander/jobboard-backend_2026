package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/study/jobboard/employer-service/internal/authtoken"
	"github.com/study/jobboard/employer-service/internal/model"
	"github.com/study/jobboard/employer-service/internal/service"
)

type employerService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (model.EmployerProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.EmployerProfile, error)
	GetCompanyPublic(ctx context.Context, employerID uuid.UUID) (model.CompanyPublic, error)
}

// EmployerHandler exposes HTTP endpoints for employer cabinet and public company info.
type EmployerHandler struct {
	svc    employerService
	tokens *authtoken.Manager
}

func NewEmployerHandler(svc employerService, tokens *authtoken.Manager) *EmployerHandler {
	return &EmployerHandler{svc: svc, tokens: tokens}
}

// GetProfile godoc
// @Summary      Get employer company profile
// @Tags         employer
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.EmployerProfile
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/employer/profile [get]
func (h *EmployerHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	profile, err := h.svc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary      Update employer company profile
// @Description  Updates company data and publishes employer.updated to RabbitMQ.
// @Tags         employer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.UpdateProfileRequest  true  "Company payload"
// @Success      200   {object}  model.EmployerProfile
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Router       /api/v1/employer/profile [put]
func (h *EmployerHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

// GetCompany godoc
// @Summary      Get public company info
// @Description  Used by vacancy-service for vacancy details page
// @Tags         companies
// @Produce      json
// @Param        id   path      string  true  "Employer profile id"
// @Success      200  {object}  model.CompanyPublic
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/employer/companies/{id} [get]
func (h *EmployerHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid company id")
		return
	}

	company, err := h.svc.GetCompanyPublic(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, company)
}

// AuthMiddleware validates Bearer JWT and requires employer role.
func (h *EmployerHandler) AuthMiddleware(next http.Handler) http.Handler {
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
		if claims.Role != authtoken.RoleEmployer {
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
