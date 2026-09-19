package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/study/jobboard/auth-service/internal/model"
)

var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, role string) (model.User, error) {
	const q = `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, created_at, updated_at
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, email, passwordHash, role).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

// GetByEmail finds a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	const q = `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	return r.scanOne(ctx, q, email)
}

// GetByID finds a user by id
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	const q = `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	return r.scanOne(ctx, q, id)
}

func (r *UserRepository) scanOne(ctx context.Context, q string, arg any) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, q, arg).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}
