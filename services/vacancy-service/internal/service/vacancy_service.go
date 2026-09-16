package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/study/jobboard/vacancy-service/internal/client"
	"github.com/study/jobboard/vacancy-service/internal/events"
	"github.com/study/jobboard/vacancy-service/internal/model"
	"github.com/study/jobboard/vacancy-service/internal/repository"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
	ErrForbiddenRole = errors.New("only employer role is allowed")
	ErrUpstream      = errors.New("employer-service unavailable")
)

// VacancyService contains vacancy CRUD / search / details logic.
type VacancyService struct {
	repo      *repository.VacancyRepository
	employer  *client.EmployerClient
	publisher events.Publisher
}

func NewVacancyService(repo *repository.VacancyRepository, employer *client.EmployerClient, publisher events.Publisher) *VacancyService {
	return &VacancyService{repo: repo, employer: employer, publisher: publisher}
}

func (s *VacancyService) Create(ctx context.Context, bearerToken string, req model.CreateVacancyRequest) (model.Vacancy, error) {
	if err := validateCreate(req); err != nil {
		return model.Vacancy{}, err
	}

	employerID, err := s.employer.GetMyEmployerID(ctx, bearerToken)
	if err != nil {
		return model.Vacancy{}, ErrUpstream
	}

	v, err := s.repo.Create(ctx, employerID, req)
	if err != nil {
		return model.Vacancy{}, err
	}
	_ = s.publisher.PublishVacancyEvent(ctx, "vacancy.created", v.ID, v.EmployerID)
	return v, nil
}

func (s *VacancyService) Update(ctx context.Context, bearerToken string, id uuid.UUID, req model.UpdateVacancyRequest) (model.Vacancy, error) {
	if err := validateUpdate(req); err != nil {
		return model.Vacancy{}, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Vacancy{}, ErrNotFound
	}
	if err != nil {
		return model.Vacancy{}, err
	}

	employerID, err := s.employer.GetMyEmployerID(ctx, bearerToken)
	if err != nil {
		return model.Vacancy{}, ErrUpstream
	}
	if existing.EmployerID != employerID {
		return model.Vacancy{}, ErrForbidden
	}

	v, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return model.Vacancy{}, err
	}
	_ = s.publisher.PublishVacancyEvent(ctx, "vacancy.updated", v.ID, v.EmployerID)
	return v, nil
}

func (s *VacancyService) Delete(ctx context.Context, bearerToken string, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	employerID, err := s.employer.GetMyEmployerID(ctx, bearerToken)
	if err != nil {
		return ErrUpstream
	}
	if existing.EmployerID != employerID {
		return ErrForbidden
	}

	v, err := s.repo.Deactivate(ctx, id)
	if err != nil {
		return err
	}
	_ = s.publisher.PublishVacancyEvent(ctx, "vacancy.deactivated", v.ID, v.EmployerID)
	return nil
}

func (s *VacancyService) Search(ctx context.Context, f model.SearchFilter) ([]model.Vacancy, error) {
	return s.repo.Search(ctx, f)
}

func (s *VacancyService) GetDetails(ctx context.Context, id uuid.UUID) (model.VacancyDetails, error) {
	v, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.VacancyDetails{}, ErrNotFound
	}
	if err != nil {
		return model.VacancyDetails{}, err
	}

	company, err := s.employer.GetCompany(ctx, v.EmployerID)
	if err != nil {
		// Fallback to local snapshot from employer.updated events.
		log.Printf("employer-service company fetch failed, trying snapshot: %v", err)
		snap, snapErr := s.repo.GetCompanySnapshot(ctx, v.EmployerID)
		if snapErr != nil {
			company = model.CompanyInfo{ID: v.EmployerID, CompanyName: "Unknown company"}
		} else {
			company = snap
		}
	}

	return model.VacancyDetails{Vacancy: v, Company: company}, nil
}

// HandleEmployerUpdated stores a local company snapshot.
func (s *VacancyService) HandleEmployerUpdated(ctx context.Context, evt model.EmployerUpdatedEvent) error {
	return s.repo.UpsertCompanySnapshot(ctx, evt.EmployerID, evt.CompanyName, evt.City)
}

func validateCreate(req model.CreateVacancyRequest) error {
	if strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Description) == "" ||
		strings.TrimSpace(req.Requirements) == "" ||
		strings.TrimSpace(req.Industry) == "" ||
		req.ExperienceYears < 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateUpdate(req model.UpdateVacancyRequest) error {
	if strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Description) == "" ||
		strings.TrimSpace(req.Requirements) == "" ||
		strings.TrimSpace(req.Industry) == "" ||
		req.ExperienceYears < 0 {
		return ErrInvalidInput
	}
	return nil
}
