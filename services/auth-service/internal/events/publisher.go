package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

type UserCreated struct {
	Event      string    `json:"event"`
	UserID     uuid.UUID `json:"user_id"`
	Role       string    `json:"role"`
	Email      string    `json:"email"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Publisher interface {
	PublishUserCreated(ctx context.Context, userID uuid.UUID, role, email string) error
	Close() error
}

type StubPublisher struct{}

func NewStubPublisher() *StubPublisher {
	return &StubPublisher{}
}

func (p *StubPublisher) PublishUserCreated(_ context.Context, userID uuid.UUID, role, email string) error {
	payload := UserCreated{
		Event:      "user.created",
		UserID:     userID,
		Role:       role,
		Email:      email,
		OccurredAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Printf("[events-stub] exchange=jobboard routing_key=user.created body=%s", string(raw))
	return nil
}

func (p *StubPublisher) Close() error {
	return nil
}
