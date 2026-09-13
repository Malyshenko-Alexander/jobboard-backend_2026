package model

import (
	"time"

	"github.com/google/uuid"
)

// EmployerProfile is employer_db.employer_profiles row.
type EmployerProfile struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	CompanyName string    `json:"company_name"`
	Description *string   `json:"description"`
	Website     *string   `json:"website"`
	City        *string   `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CompanyPublic is a public DTO for vacancy details page.
type CompanyPublic struct {
	ID          uuid.UUID `json:"id"`
	CompanyName string    `json:"company_name"`
	Description *string   `json:"description"`
	Website     *string   `json:"website"`
	City        *string   `json:"city"`
}

// ToPublic maps profile to public company view.
func (p EmployerProfile) ToPublic() CompanyPublic {
	return CompanyPublic{
		ID:          p.ID,
		CompanyName: p.CompanyName,
		Description: p.Description,
		Website:     p.Website,
		City:        p.City,
	}
}

// UpdateProfileRequest is the body for PUT /employer/profile.
type UpdateProfileRequest struct {
	CompanyName string  `json:"company_name" example:"Acme LLC"`
	Description *string `json:"description" example:"We build cool stuff"`
	Website     *string `json:"website" example:"https://acme.example"`
	City        *string `json:"city" example:"Moscow"`
}

// UserCreatedEvent matches auth-service payload for user.created.
type UserCreatedEvent struct {
	Event      string    `json:"event"`
	UserID     uuid.UUID `json:"user_id"`
	Role       string    `json:"role"`
	Email      string    `json:"email"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EmployerUpdatedEvent is published after company profile changes.
type EmployerUpdatedEvent struct {
	Event       string    `json:"event"`
	EmployerID  uuid.UUID `json:"employer_id"`
	CompanyName string    `json:"company_name"`
	City        *string   `json:"city"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// ErrorResponse is a simple API error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"profile not found"`
}
