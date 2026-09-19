package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleApplicant = "applicant"
	RoleEmployer  = "employer"
)

// User is the auth_db.users row
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
	Role     string `json:"role" example:"applicant" enums:"applicant,employer"`
}

// LoginRequest
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
}

// AuthResponse
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	Role      string    `json:"role" example:"applicant"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse maps User to UserResponse
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// ErrorResponse
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
