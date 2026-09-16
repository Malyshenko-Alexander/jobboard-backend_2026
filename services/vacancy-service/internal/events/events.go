package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/study/jobboard/vacancy-service/internal/model"
)

// Publisher sends vacancy lifecycle events.
type Publisher interface {
	PublishVacancyEvent(ctx context.Context, eventName string, vacancyID, employerID uuid.UUID) error
	Close() error
}

type StubPublisher struct{}

func NewStubPublisher() *StubPublisher {
	return &StubPublisher{}
}

func (p *StubPublisher) PublishVacancyEvent(_ context.Context, eventName string, vacancyID, employerID uuid.UUID) error {
	payload := model.VacancyEvent{
		Event:      eventName,
		VacancyID:  vacancyID,
		EmployerID: employerID,
		OccurredAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	log.Printf("[events-stub] exchange=jobboard routing_key=%s body=%s", eventName, string(raw))
	return nil
}

func (p *StubPublisher) Close() error {
	return nil
}

// EmployerUpdatedProcessor is implemented by VacancyService.
type EmployerUpdatedProcessor interface {
	HandleEmployerUpdated(ctx context.Context, evt model.EmployerUpdatedEvent) error
}

type StubConsumer struct {
	processor EmployerUpdatedProcessor
}

func NewStubConsumer(processor EmployerUpdatedProcessor) *StubConsumer {
	return &StubConsumer{processor: processor}
}

func (c *StubConsumer) Start(_ context.Context) error {
	log.Printf("[events-stub] vacancy consumer ready (queue=vacancy.employer_updated), RabbitMQ not connected yet")
	return nil
}

func (c *StubConsumer) Close() error {
	return nil
}

func (c *StubConsumer) HandleEmployerUpdated(ctx context.Context, raw []byte) error {
	var evt model.EmployerUpdatedEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	log.Printf("[events-stub] received employer.updated employer_id=%s", evt.EmployerID)
	return c.processor.HandleEmployerUpdated(ctx, evt)
}
