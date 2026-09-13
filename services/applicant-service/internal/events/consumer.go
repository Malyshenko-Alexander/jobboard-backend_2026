package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/study/jobboard/applicant-service/internal/model"
	"github.com/study/jobboard/applicant-service/internal/service"
)

// Consumer receives domain events from other microservices.
type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type StubConsumer struct {
	svc *service.ApplicantService
}

func NewStubConsumer(svc *service.ApplicantService) *StubConsumer {
	return &StubConsumer{svc: svc}
}

func (c *StubConsumer) Start(_ context.Context) error {
	log.Printf("[events-stub] applicant consumer ready (queue=applicant.user_created), RabbitMQ not connected yet")
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
	return c.svc.HandleUserCreated(ctx, evt)
}
