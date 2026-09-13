package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/study/jobboard/applicant-service/internal/model"
)

var ErrNotFound = errors.New("not found")

// ProfileRepository talks to applicant_profiles.
type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// CreateEmpty inserts a minimal profile for a new applicant user.
func (r *ProfileRepository) CreateEmpty(ctx context.Context, userID uuid.UUID) (model.ApplicantProfile, error) {
	const q = `
		INSERT INTO applicant_profiles (user_id, full_name)
		VALUES ($1, '')
		ON CONFLICT (user_id) DO UPDATE SET updated_at = applicant_profiles.updated_at
		RETURNING id, user_id, full_name, phone, city, about, created_at, updated_at
	`
	return r.scanProfile(r.db.QueryRow(ctx, q, userID))
}

// GetByUserID finds profile by auth user id.
func (r *ProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (model.ApplicantProfile, error) {
	const q = `
		SELECT id, user_id, full_name, phone, city, about, created_at, updated_at
		FROM applicant_profiles
		WHERE user_id = $1
	`
	p, err := r.scanProfile(r.db.QueryRow(ctx, q, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ApplicantProfile{}, ErrNotFound
	}
	return p, err
}

// UpdateByUserID updates profile fields for the given user.
func (r *ProfileRepository) UpdateByUserID(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.ApplicantProfile, error) {
	const q = `
		UPDATE applicant_profiles
		SET full_name = $2,
		    phone = $3,
		    city = $4,
		    about = $5,
		    updated_at = NOW()
		WHERE user_id = $1
		RETURNING id, user_id, full_name, phone, city, about, created_at, updated_at
	`
	p, err := r.scanProfile(r.db.QueryRow(ctx, q, userID, req.FullName, req.Phone, req.City, req.About))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ApplicantProfile{}, ErrNotFound
	}
	return p, err
}

func (r *ProfileRepository) scanProfile(row pgx.Row) (model.ApplicantProfile, error) {
	var p model.ApplicantProfile
	err := row.Scan(&p.ID, &p.UserID, &p.FullName, &p.Phone, &p.City, &p.About, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// ResumeRepository talks to resumes.
type ResumeRepository struct {
	db *pgxpool.Pool
}

func NewResumeRepository(db *pgxpool.Pool) *ResumeRepository {
	return &ResumeRepository{db: db}
}

// GetActiveByApplicantID returns the latest active resume.
func (r *ResumeRepository) GetActiveByApplicantID(ctx context.Context, applicantID uuid.UUID) (model.Resume, error) {
	const q = `
		SELECT id, applicant_id, title, summary, skills, experience_years,
		       education, desired_salary, is_active, created_at, updated_at
		FROM resumes
		WHERE applicant_id = $1 AND is_active = TRUE
		ORDER BY updated_at DESC
		LIMIT 1
	`
	res, err := r.scanResume(r.db.QueryRow(ctx, q, applicantID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Resume{}, ErrNotFound
	}
	return res, err
}

// UpsertActive creates a resume or updates the current active one.
func (r *ResumeRepository) UpsertActive(ctx context.Context, applicantID uuid.UUID, req model.UpsertResumeRequest) (model.Resume, error) {
	existing, err := r.GetActiveByApplicantID(ctx, applicantID)
	if errors.Is(err, ErrNotFound) {
		return r.insert(ctx, applicantID, req)
	}
	if err != nil {
		return model.Resume{}, err
	}
	return r.update(ctx, existing.ID, req)
}

func (r *ResumeRepository) insert(ctx context.Context, applicantID uuid.UUID, req model.UpsertResumeRequest) (model.Resume, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	const q = `
		INSERT INTO resumes (
			applicant_id, title, summary, skills, experience_years,
			education, desired_salary, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, applicant_id, title, summary, skills, experience_years,
		          education, desired_salary, is_active, created_at, updated_at
	`
	return r.scanResume(r.db.QueryRow(
		ctx, q,
		applicantID, req.Title, req.Summary, req.Skills, req.ExperienceYears,
		req.Education, req.DesiredSalary, isActive,
	))
}

func (r *ResumeRepository) update(ctx context.Context, id uuid.UUID, req model.UpsertResumeRequest) (model.Resume, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	const q = `
		UPDATE resumes
		SET title = $2,
		    summary = $3,
		    skills = $4,
		    experience_years = $5,
		    education = $6,
		    desired_salary = $7,
		    is_active = $8,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, applicant_id, title, summary, skills, experience_years,
		          education, desired_salary, is_active, created_at, updated_at
	`
	return r.scanResume(r.db.QueryRow(
		ctx, q,
		id, req.Title, req.Summary, req.Skills, req.ExperienceYears,
		req.Education, req.DesiredSalary, isActive,
	))
}

func (r *ResumeRepository) scanResume(row pgx.Row) (model.Resume, error) {
	var res model.Resume
	err := row.Scan(
		&res.ID, &res.ApplicantID, &res.Title, &res.Summary, &res.Skills, &res.ExperienceYears,
		&res.Education, &res.DesiredSalary, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)
	return res, err
}
