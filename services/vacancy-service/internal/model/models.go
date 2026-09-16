package model

import (
	"time"

	"github.com/google/uuid"
)

// Vacancy is vacancy_db.vacancies row.
type Vacancy struct {
	ID              uuid.UUID `json:"id"`
	EmployerID      uuid.UUID `json:"employer_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Requirements    string    `json:"requirements"`
	Industry        string    `json:"industry"`
	SalaryFrom      *int      `json:"salary_from"`
	SalaryTo        *int      `json:"salary_to"`
	ExperienceYears int       `json:"experience_years"`
	City            *string   `json:"city"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateVacancyRequest is the body for POST /vacancies.
type CreateVacancyRequest struct {
	Title           string  `json:"title" example:"Go Developer"`
	Description     string  `json:"description" example:"Build microservices"`
	Requirements    string  `json:"requirements" example:"Go, PostgreSQL"`
	Industry        string  `json:"industry" example:"IT"`
	SalaryFrom      *int    `json:"salary_from" example:"150000"`
	SalaryTo        *int    `json:"salary_to" example:"250000"`
	ExperienceYears int     `json:"experience_years" example:"2"`
	City            *string `json:"city" example:"Moscow"`
}

// UpdateVacancyRequest is the body for PUT /vacancies/{id}.
type UpdateVacancyRequest struct {
	Title           string  `json:"title" example:"Senior Go Developer"`
	Description     string  `json:"description" example:"Lead backend work"`
	Requirements    string  `json:"requirements" example:"Go, RabbitMQ"`
	Industry        string  `json:"industry" example:"IT"`
	SalaryFrom      *int    `json:"salary_from" example:"200000"`
	SalaryTo        *int    `json:"salary_to" example:"300000"`
	ExperienceYears int     `json:"experience_years" example:"3"`
	City            *string `json:"city" example:"Moscow"`
	IsActive        *bool   `json:"is_active" example:"true"`
}

// SearchFilter holds query params for vacancy search.
type SearchFilter struct {
	Industry        string
	SalaryFrom      *int
	SalaryTo        *int
	ExperienceYears *int
}

// CompanyInfo is public company data from employer-service.
type CompanyInfo struct {
	ID          uuid.UUID `json:"id"`
	CompanyName string    `json:"company_name"`
	Description *string   `json:"description,omitempty"`
	Website     *string   `json:"website,omitempty"`
	City        *string   `json:"city,omitempty"`
}

// VacancyDetails is vacancy + company block for details page.
type VacancyDetails struct {
	Vacancy Vacancy     `json:"vacancy"`
	Company CompanyInfo `json:"company"`
}

// EmployerUpdatedEvent matches employer-service payload.
type EmployerUpdatedEvent struct {
	Event       string    `json:"event"`
	EmployerID  uuid.UUID `json:"employer_id"`
	CompanyName string    `json:"company_name"`
	City        *string   `json:"city"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// VacancyEvent is published on vacancy lifecycle changes.
type VacancyEvent struct {
	Event      string    `json:"event"`
	VacancyID  uuid.UUID `json:"vacancy_id"`
	EmployerID uuid.UUID `json:"employer_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ErrorResponse is a simple API error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"vacancy not found"`
}
