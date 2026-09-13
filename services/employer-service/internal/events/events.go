package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/study/jobboard/employer-service/internal/model"
)

type Publisher interface {
	PublishEmployerUpdated(ctx context.Context, employerID uuid.UUID, companyName string, city *string) error
	Close() error
}

type StubPublisher struct{}

func NewStubPublisher() *StubPublisher {
	return &StubPublisher{}
}

func (p *StubPublisher) PublishEmployerUpdated(_ context.Context, employerID uuid.UUID, companyName string, city *string) error {
	payload := model.EmployerUpdatedEvent{
		Event:       "employer.updated",
		EmployerID:  employerID,
		CompanyName: companyName,
		City:        city,
		OccurredAt:  time.Now().UTC(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	log.Printf("[events-stub] exchange=jobboard routing_key=employer.updated body=%s", string(raw))
	return nil
}

func (p *StubPublisher) Close() error {
	return nil
}

// UserCreatedProcessor is implemented by EmployerService.
type UserCreatedProcessor interface {
	HandleUserCreated(ctx context.Context, evt model.UserCreatedEvent) error
}

type StubConsumer struct {
	processor UserCreatedProcessor
}

func NewStubConsumer(processor UserCreatedProcessor) *StubConsumer {
	return &StubConsumer{processor: processor}
}

func (c *StubConsumer) Start(_ context.Context) error {
	log.Printf("[events-stub] employer consumer ready (queue=employer.user_created), RabbitMQ not connected yet")
	return nil
}

func (c *StubConsumer) Close() error {
	return nil
}

// HandleUserCreated processes user.created payload.
func (c *StubConsumer) HandleUserCreated(ctx context.Context, raw []byte) error {
	var evt model.UserCreatedEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	log.Printf("[events-stub] received user.created user_id=%s role=%s", evt.UserID, evt.Role)
	return c.processor.HandleUserCreated(ctx, evt)
}
