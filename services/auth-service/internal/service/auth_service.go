package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/study/jobboard/auth-service/internal/authtoken"
	"github.com/study/jobboard/auth-service/internal/events"
	"github.com/study/jobboard/auth-service/internal/model"
	"github.com/study/jobboard/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidRole        = errors.New("role must be applicant or employer")
)

type AuthService struct {
	users     *repository.UserRepository
	tokens    *authtoken.Manager
	publisher events.Publisher
}

func NewAuthService(users *repository.UserRepository, tokens *authtoken.Manager, publisher events.Publisher) *AuthService {
	return &AuthService{users: users, tokens: tokens, publisher: publisher}
}

// Register creates a user, publishes user.created, returns JWT.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := strings.TrimSpace(req.Password)
	role := strings.TrimSpace(req.Role)

	if email == "" || len(password) < 6 {
		return model.AuthResponse{}, ErrInvalidInput
	}
	if role != model.RoleApplicant && role != model.RoleEmployer {
		return model.AuthResponse{}, ErrInvalidRole
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.AuthResponse{}, err
	}

	user, err := s.users.Create(ctx, email, string(hash), role)
	if err != nil {
		if isUniqueViolation(err) {
			return model.AuthResponse{}, ErrEmailTaken
		}
		return model.AuthResponse{}, err
	}

	// Notify other services to create an empty profile
	_ = s.publisher.PublishUserCreated(ctx, user.ID, user.Role, user.Email)

	token, err := s.tokens.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{Token: token, User: user.ToResponse()}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return model.AuthResponse{}, ErrInvalidInput
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AuthResponse{}, ErrInvalidCredentials
		}
		return model.AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{Token: token, User: user.ToResponse()}, nil
}

// Me returns the current user by id from JWT
func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (model.UserResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.UserResponse{}, ErrInvalidCredentials
		}
		return model.UserResponse{}, err
	}
	return user.ToResponse(), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
