package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/study/jobboard/auth-service/internal/authtoken"
	"github.com/study/jobboard/auth-service/internal/model"
	"github.com/study/jobboard/auth-service/internal/service"
)

type authService interface {
	Register(ctx context.Context, req model.RegisterRequest) (model.AuthResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error)
	Me(ctx context.Context, userID uuid.UUID) (model.UserResponse, error)
}

// AuthHandler exposes HTTP endpoints for auth-service.
type AuthHandler struct {
	svc    authService
	tokens *authtoken.Manager
}

func NewAuthHandler(svc authService, tokens *authtoken.Manager) *AuthHandler {
	return &AuthHandler{svc: svc, tokens: tokens}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates applicant or employer account and returns JWT. Publishes user.created event (stub).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest  true  "Register payload"
// @Success      201   {object}  model.AuthResponse
// @Failure      400   {object}  model.ErrorResponse
// @Failure      409   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrInvalidRole):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrEmailTaken):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login
// @Description  Validates credentials and returns JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.LoginRequest  true  "Login payload"
// @Success      200   {object}  model.AuthResponse
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Me godoc
// @Summary      Current user
// @Description  Returns the authenticated user from JWT
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.UserResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.Me(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// AuthMiddleware validates Bearer JWT and puts claims into context.
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
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

		ctx := ContextWithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}
