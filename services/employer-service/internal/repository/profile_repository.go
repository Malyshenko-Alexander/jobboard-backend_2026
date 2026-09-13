package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/study/jobboard/employer-service/internal/model"
)

var ErrNotFound = errors.New("not found")

// ProfileRepository talks to employer_profiles.
type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// CreateEmpty inserts a minimal company profile for a new employer user.
func (r *ProfileRepository) CreateEmpty(ctx context.Context, userID uuid.UUID) (model.EmployerProfile, error) {
	const q = `
		INSERT INTO employer_profiles (user_id, company_name)
		VALUES ($1, '')
		ON CONFLICT (user_id) DO UPDATE SET updated_at = employer_profiles.updated_at
		RETURNING id, user_id, company_name, description, website, city, created_at, updated_at
	`
	return r.scan(r.db.QueryRow(ctx, q, userID))
}

// GetByUserID finds profile by auth user id.
func (r *ProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (model.EmployerProfile, error) {
	const q = `
		SELECT id, user_id, company_name, description, website, city, created_at, updated_at
		FROM employer_profiles
		WHERE user_id = $1
	`
	p, err := r.scan(r.db.QueryRow(ctx, q, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.EmployerProfile{}, ErrNotFound
	}
	return p, err
}

// GetByID finds profile by employer profile id (public company lookup).
func (r *ProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (model.EmployerProfile, error) {
	const q = `
		SELECT id, user_id, company_name, description, website, city, created_at, updated_at
		FROM employer_profiles
		WHERE id = $1
	`
	p, err := r.scan(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.EmployerProfile{}, ErrNotFound
	}
	return p, err
}

// UpdateByUserID updates company fields for the given user.
func (r *ProfileRepository) UpdateByUserID(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.EmployerProfile, error) {
	const q = `
		UPDATE employer_profiles
		SET company_name = $2,
		    description = $3,
		    website = $4,
		    city = $5,
		    updated_at = NOW()
		WHERE user_id = $1
		RETURNING id, user_id, company_name, description, website, city, created_at, updated_at
	`
	p, err := r.scan(r.db.QueryRow(ctx, q, userID, req.CompanyName, req.Description, req.Website, req.City))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.EmployerProfile{}, ErrNotFound
	}
	return p, err
}

func (r *ProfileRepository) scan(row pgx.Row) (model.EmployerProfile, error) {
	var p model.EmployerProfile
	err := row.Scan(&p.ID, &p.UserID, &p.CompanyName, &p.Description, &p.Website, &p.City, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}
