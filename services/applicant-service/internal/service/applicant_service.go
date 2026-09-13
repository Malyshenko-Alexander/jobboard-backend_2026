package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/study/jobboard/applicant-service/internal/authtoken"
	"github.com/study/jobboard/applicant-service/internal/model"
	"github.com/study/jobboard/applicant-service/internal/repository"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
	ErrForbiddenRole = errors.New("only applicant role is allowed")
)

// ApplicantService contains profile / resume logic.
type ApplicantService struct {
	profiles *repository.ProfileRepository
	resumes  *repository.ResumeRepository
}

func NewApplicantService(profiles *repository.ProfileRepository, resumes *repository.ResumeRepository) *ApplicantService {
	return &ApplicantService{profiles: profiles, resumes: resumes}
}

// HandleUserCreated creates an empty profile when auth publishes user.created.
func (s *ApplicantService) HandleUserCreated(ctx context.Context, evt model.UserCreatedEvent) error {
	if evt.Role != authtoken.RoleApplicant {
		// Ignore employer registrations — not our concern.
		return nil
	}
	_, err := s.profiles.CreateEmpty(ctx, evt.UserID)
	return err
}

// GetProfile returns profile for the authenticated applicant.
func (s *ApplicantService) GetProfile(ctx context.Context, userID uuid.UUID) (model.ApplicantProfile, error) {
	p, err := s.profiles.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.ApplicantProfile{}, ErrNotFound
	}
	return p, err
}

// UpdateProfile updates profile fields.
func (s *ApplicantService) UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.ApplicantProfile, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		return model.ApplicantProfile{}, ErrInvalidInput
	}

	p, err := s.profiles.UpdateByUserID(ctx, userID, req)
	if errors.Is(err, repository.ErrNotFound) {
		return model.ApplicantProfile{}, ErrNotFound
	}
	return p, err
}

// GetResume returns active resume for the user.
func (s *ApplicantService) GetResume(ctx context.Context, userID uuid.UUID) (model.Resume, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Resume{}, ErrNotFound
	}
	if err != nil {
		return model.Resume{}, err
	}

	resume, err := s.resumes.GetActiveByApplicantID(ctx, profile.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Resume{}, ErrNotFound
	}
	return resume, err
}

// UpsertResume creates or updates the active resume.
func (s *ApplicantService) UpsertResume(ctx context.Context, userID uuid.UUID, req model.UpsertResumeRequest) (model.Resume, error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || req.ExperienceYears < 0 {
		return model.Resume{}, ErrInvalidInput
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Resume{}, ErrNotFound
	}
	if err != nil {
		return model.Resume{}, err
	}

	return s.resumes.UpsertActive(ctx, profile.ID, req)
}
