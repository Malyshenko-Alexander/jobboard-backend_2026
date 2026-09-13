package model

import (
	"time"

	"github.com/google/uuid"
)

// ApplicantProfile is applicant_db.applicant_profiles row.
type ApplicantProfile struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FullName  string    `json:"full_name"`
	Phone     *string   `json:"phone"`
	City      *string   `json:"city"`
	About     *string   `json:"about"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateProfileRequest is the body for PUT /applicant/profile.
type UpdateProfileRequest struct {
	FullName string  `json:"full_name" example:"Ivan Ivanov"`
	Phone    *string `json:"phone" example:"+79001234567"`
	City     *string `json:"city" example:"Moscow"`
	About    *string `json:"about" example:"Backend developer"`
}

// Resume is applicant_db.resumes row.
type Resume struct {
	ID              uuid.UUID `json:"id"`
	ApplicantID     uuid.UUID `json:"applicant_id"`
	Title           string    `json:"title"`
	Summary         *string   `json:"summary"`
	Skills          *string   `json:"skills"`
	ExperienceYears int       `json:"experience_years"`
	Education       *string   `json:"education"`
	DesiredSalary   *int      `json:"desired_salary"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UpsertResumeRequest is the body for PUT /applicant/resume.
type UpsertResumeRequest struct {
	Title           string  `json:"title" example:"Go Developer"`
	Summary         *string `json:"summary" example:"2 years of backend experience"`
	Skills          *string `json:"skills" example:"Go, PostgreSQL, RabbitMQ"`
	ExperienceYears int     `json:"experience_years" example:"2"`
	Education       *string `json:"education" example:"BS Computer Science"`
	DesiredSalary   *int    `json:"desired_salary" example:"150000"`
	IsActive        *bool   `json:"is_active" example:"true"`
}

// UserCreatedEvent matches auth-service payload for user.created.
type UserCreatedEvent struct {
	Event      string    `json:"event"`
	UserID     uuid.UUID `json:"user_id"`
	Role       string    `json:"role"`
	Email      string    `json:"email"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ErrorResponse is a simple API error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"profile not found"`
}
