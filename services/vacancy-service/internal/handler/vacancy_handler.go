package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/study/jobboard/vacancy-service/internal/authtoken"
	"github.com/study/jobboard/vacancy-service/internal/model"
	"github.com/study/jobboard/vacancy-service/internal/service"
)

type vacancyService interface {
	Create(ctx context.Context, bearerToken string, req model.CreateVacancyRequest) (model.Vacancy, error)
	Update(ctx context.Context, bearerToken string, id uuid.UUID, req model.UpdateVacancyRequest) (model.Vacancy, error)
	Delete(ctx context.Context, bearerToken string, id uuid.UUID) error
	Search(ctx context.Context, f model.SearchFilter) ([]model.Vacancy, error)
	GetDetails(ctx context.Context, id uuid.UUID) (model.VacancyDetails, error)
}

type employerUpdatedHandler interface {
	HandleEmployerUpdated(ctx context.Context, raw []byte) error
}

// VacancyHandler exposes HTTP endpoints for vacancies.
type VacancyHandler struct {
	svc      vacancyService
	tokens   *authtoken.Manager
	consumer employerUpdatedHandler
}

func NewVacancyHandler(svc vacancyService, tokens *authtoken.Manager, consumer employerUpdatedHandler) *VacancyHandler {
	return &VacancyHandler{svc: svc, tokens: tokens, consumer: consumer}
}

// ListVacancies godoc
// @Summary      Search vacancies
// @Tags         vacancies
// @Produce      json
// @Param        industry          query     string  false  "Industry filter"
// @Param        salary_from       query     int     false  "Minimum salary"
// @Param        salary_to         query     int     false  "Maximum salary"
// @Param        experience_years  query     int     false  "Max required experience (candidate years)"
// @Success      200  {array}   model.Vacancy
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/vacancies [get]
func (h *VacancyHandler) ListVacancies(w http.ResponseWriter, r *http.Request) {
	f := model.SearchFilter{Industry: strings.TrimSpace(r.URL.Query().Get("industry"))}
	if v := r.URL.Query().Get("salary_from"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid salary_from")
			return
		}
		f.SalaryFrom = &n
	}
	if v := r.URL.Query().Get("salary_to"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid salary_to")
			return
		}
		f.SalaryTo = &n
	}
	if v := r.URL.Query().Get("experience_years"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid experience_years")
			return
		}
		f.ExperienceYears = &n
	}

	list, err := h.svc.Search(r.Context(), f)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GetVacancy godoc
// @Summary      Vacancy details
// @Tags         vacancies
// @Produce      json
// @Param        id   path      string  true  "Vacancy id"
// @Success      200  {object}  model.VacancyDetails
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/vacancies/{id} [get]
func (h *VacancyHandler) GetVacancy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	details, err := h.svc.GetDetails(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, details)
}

// CreateVacancy godoc
// @Summary      Create vacancy
// @Tags         vacancies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateVacancyRequest  true  "Vacancy payload"
// @Success      201   {object}  model.Vacancy
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      502   {object}  model.ErrorResponse
// @Router       /api/v1/vacancies [post]
func (h *VacancyHandler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVacancyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	v, err := h.svc.Create(r.Context(), TokenFromContext(r.Context()), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

// UpdateVacancy godoc
// @Summary      Update own vacancy
// @Tags         vacancies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                     true  "Vacancy id"
// @Param        body  body      model.UpdateVacancyRequest  true  "Vacancy payload"
// @Success      200   {object}  model.Vacancy
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Router       /api/v1/vacancies/{id} [put]
func (h *VacancyHandler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	var req model.UpdateVacancyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	v, err := h.svc.Update(r.Context(), TokenFromContext(r.Context()), id, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// DeleteVacancy godoc
// @Summary      Deactivate own vacancy
// @Tags         vacancies
// @Security     BearerAuth
// @Param        id   path  string  true  "Vacancy id"
// @Success      204
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/vacancies/{id} [delete]
func (h *VacancyHandler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	if err := h.svc.Delete(r.Context(), TokenFromContext(r.Context()), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EmployerUpdatedHook godoc
// @Summary      Debug hook for employer.updated (prefer RabbitMQ)
// @Tags         internal
// @Accept       json
// @Param        body  body  model.EmployerUpdatedEvent  true  "employer.updated payload"
// @Success      204
// @Failure      400  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/internal/events/employer-updated [post]
func (h *VacancyHandler) EmployerUpdatedHook(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.consumer.HandleEmployerUpdated(r.Context(), raw); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to handle event")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AuthMiddleware validates Bearer JWT and requires employer role.
func (h *VacancyHandler) AuthMiddleware(next http.Handler) http.Handler {
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
		ctx = ContextWithToken(ctx, tokenStr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrForbiddenRole):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrUpstream):
		writeError(w, http.StatusBadGateway, err.Error())
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
