package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/study/jobboard/employer-service/internal/authtoken"
	"github.com/study/jobboard/employer-service/internal/events"
	"github.com/study/jobboard/employer-service/internal/model"
	"github.com/study/jobboard/employer-service/internal/repository"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
	ErrForbiddenRole = errors.New("only employer role is allowed")
)

// EmployerService contains company profile logic.
type EmployerService struct {
	profiles  *repository.ProfileRepository
	publisher events.Publisher
}

func NewEmployerService(profiles *repository.ProfileRepository, publisher events.Publisher) *EmployerService {
	return &EmployerService{profiles: profiles, publisher: publisher}
}

// HandleUserCreated creates an empty company profile for employer registrations.
func (s *EmployerService) HandleUserCreated(ctx context.Context, evt model.UserCreatedEvent) error {
	if evt.Role != authtoken.RoleEmployer {
		return nil
	}
	_, err := s.profiles.CreateEmpty(ctx, evt.UserID)
	return err
}

// GetProfile returns company profile for the authenticated employer.
func (s *EmployerService) GetProfile(ctx context.Context, userID uuid.UUID) (model.EmployerProfile, error) {
	p, err := s.profiles.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.EmployerProfile{}, ErrNotFound
	}
	return p, err
}

// UpdateProfile updates company fields and publishes employer.updated.
func (s *EmployerService) UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.EmployerProfile, error) {
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	if req.CompanyName == "" {
		return model.EmployerProfile{}, ErrInvalidInput
	}

	p, err := s.profiles.UpdateByUserID(ctx, userID, req)
	if errors.Is(err, repository.ErrNotFound) {
		return model.EmployerProfile{}, ErrNotFound
	}
	if err != nil {
		return model.EmployerProfile{}, err
	}

	_ = s.publisher.PublishEmployerUpdated(ctx, p.ID, p.CompanyName, p.City)
	return p, nil
}

// GetCompanyPublic returns public company info by employer profile id.
func (s *EmployerService) GetCompanyPublic(ctx context.Context, employerID uuid.UUID) (model.CompanyPublic, error) {
	p, err := s.profiles.GetByID(ctx, employerID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.CompanyPublic{}, ErrNotFound
	}
	if err != nil {
		return model.CompanyPublic{}, err
	}
	return p.ToPublic(), nil
}
