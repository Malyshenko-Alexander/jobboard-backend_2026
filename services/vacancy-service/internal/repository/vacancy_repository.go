package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/study/jobboard/vacancy-service/internal/model"
)

var ErrNotFound = errors.New("not found")

// VacancyRepository talks to vacancies and company_snapshots.
type VacancyRepository struct {
	db *pgxpool.Pool
}

func NewVacancyRepository(db *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{db: db}
}

func (r *VacancyRepository) Create(ctx context.Context, employerID uuid.UUID, req model.CreateVacancyRequest) (model.Vacancy, error) {
	const q = `
		INSERT INTO vacancies (
			employer_id, title, description, requirements, industry,
			salary_from, salary_to, experience_years, city
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, employer_id, title, description, requirements, industry,
		          salary_from, salary_to, experience_years, city, is_active, created_at, updated_at
	`
	return r.scan(r.db.QueryRow(
		ctx, q,
		employerID, req.Title, req.Description, req.Requirements, req.Industry,
		req.SalaryFrom, req.SalaryTo, req.ExperienceYears, req.City,
	))
}

func (r *VacancyRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Vacancy, error) {
	const q = `
		SELECT id, employer_id, title, description, requirements, industry,
		       salary_from, salary_to, experience_years, city, is_active, created_at, updated_at
		FROM vacancies
		WHERE id = $1
	`
	v, err := r.scan(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Vacancy{}, ErrNotFound
	}
	return v, err
}

func (r *VacancyRepository) Update(ctx context.Context, id uuid.UUID, req model.UpdateVacancyRequest) (model.Vacancy, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	const q = `
		UPDATE vacancies
		SET title = $2,
		    description = $3,
		    requirements = $4,
		    industry = $5,
		    salary_from = $6,
		    salary_to = $7,
		    experience_years = $8,
		    city = $9,
		    is_active = $10,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, employer_id, title, description, requirements, industry,
		          salary_from, salary_to, experience_years, city, is_active, created_at, updated_at
	`
	v, err := r.scan(r.db.QueryRow(
		ctx, q,
		id, req.Title, req.Description, req.Requirements, req.Industry,
		req.SalaryFrom, req.SalaryTo, req.ExperienceYears, req.City, isActive,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Vacancy{}, ErrNotFound
	}
	return v, err
}

func (r *VacancyRepository) Deactivate(ctx context.Context, id uuid.UUID) (model.Vacancy, error) {
	const q = `
		UPDATE vacancies
		SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1
		RETURNING id, employer_id, title, description, requirements, industry,
		          salary_from, salary_to, experience_years, city, is_active, created_at, updated_at
	`
	v, err := r.scan(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Vacancy{}, ErrNotFound
	}
	return v, err
}

// Search returns active vacancies matching optional filters.
func (r *VacancyRepository) Search(ctx context.Context, f model.SearchFilter) ([]model.Vacancy, error) {
	var (
		b    strings.Builder
		args []any
		n    = 1
	)

	b.WriteString(`
		SELECT id, employer_id, title, description, requirements, industry,
		       salary_from, salary_to, experience_years, city, is_active, created_at, updated_at
		FROM vacancies
		WHERE is_active = TRUE
	`)

	if f.Industry != "" {
		b.WriteString(fmt.Sprintf(" AND industry ILIKE $%d", n))
		args = append(args, f.Industry)
		n++
	}
	if f.SalaryFrom != nil {
		// vacancy upper bound should be >= requested min (or null salary treated as unknown / skip)
		b.WriteString(fmt.Sprintf(" AND (salary_to IS NULL OR salary_to >= $%d)", n))
		args = append(args, *f.SalaryFrom)
		n++
	}
	if f.SalaryTo != nil {
		b.WriteString(fmt.Sprintf(" AND (salary_from IS NULL OR salary_from <= $%d)", n))
		args = append(args, *f.SalaryTo)
		n++
	}
	if f.ExperienceYears != nil {
		b.WriteString(fmt.Sprintf(" AND experience_years <= $%d", n))
		args = append(args, *f.ExperienceYears)
		n++
	}

	b.WriteString(" ORDER BY created_at DESC")

	rows, err := r.db.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Vacancy
	for rows.Next() {
		v, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	if list == nil {
		list = []model.Vacancy{}
	}
	return list, rows.Err()
}

func (r *VacancyRepository) UpsertCompanySnapshot(ctx context.Context, employerID uuid.UUID, companyName string, city *string) error {
	const q = `
		INSERT INTO company_snapshots (employer_id, company_name, city, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (employer_id) DO UPDATE
		SET company_name = EXCLUDED.company_name,
		    city = EXCLUDED.city,
		    updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, q, employerID, companyName, city)
	return err
}

func (r *VacancyRepository) GetCompanySnapshot(ctx context.Context, employerID uuid.UUID) (model.CompanyInfo, error) {
	const q = `
		SELECT employer_id, company_name, city
		FROM company_snapshots
		WHERE employer_id = $1
	`
	var info model.CompanyInfo
	err := r.db.QueryRow(ctx, q, employerID).Scan(&info.ID, &info.CompanyName, &info.City)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CompanyInfo{}, ErrNotFound
	}
	return info, err
}

type scannable interface {
	Scan(dest ...any) error
}

func (r *VacancyRepository) scan(row scannable) (model.Vacancy, error) {
	var v model.Vacancy
	err := row.Scan(
		&v.ID, &v.EmployerID, &v.Title, &v.Description, &v.Requirements, &v.Industry,
		&v.SalaryFrom, &v.SalaryTo, &v.ExperienceYears, &v.City, &v.IsActive, &v.CreatedAt, &v.UpdatedAt,
	)
	return v, err
}
