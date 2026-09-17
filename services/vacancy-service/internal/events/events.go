package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	"github.com/study/jobboard/vacancy-service/internal/model"
)

const (
	exchangeName           = "jobboard"
	routingEmployerUpdated = "employer.updated"
	queueEmployerUpdated   = "vacancy.employer_updated"
)

// Publisher sends vacancy lifecycle events.
type Publisher interface {
	PublishVacancyEvent(ctx context.Context, eventName string, vacancyID, employerID uuid.UUID) error
	Close() error
}

// RabbitPublisher publishes vacancy.* events.
type RabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitPublisher(url string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbit dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &RabbitPublisher{conn: conn, ch: ch}, nil
}

func (p *RabbitPublisher) PublishVacancyEvent(ctx context.Context, eventName string, vacancyID, employerID uuid.UUID) error {
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

	return p.ch.PublishWithContext(ctx, exchangeName, eventName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         raw,
	})
}

func (p *RabbitPublisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// EmployerUpdatedProcessor is implemented by VacancyService.
type EmployerUpdatedProcessor interface {
	HandleEmployerUpdated(ctx context.Context, evt model.EmployerUpdatedEvent) error
}

// RabbitConsumer listens for employer.updated.
type RabbitConsumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	processor EmployerUpdatedProcessor
}

func NewRabbitConsumer(url string, processor EmployerUpdatedProcessor) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbit dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	_, err = ch.QueueDeclare(queueEmployerUpdated, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(queueEmployerUpdated, routingEmployerUpdated, exchangeName, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	return &RabbitConsumer{conn: conn, ch: ch, processor: processor}, nil
}

func (c *RabbitConsumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume(queueEmployerUpdated, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		log.Printf("rabbit consumer started queue=%s", queueEmployerUpdated)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				if err := c.HandleEmployerUpdated(ctx, d.Body); err != nil {
					log.Printf("employer.updated handle error: %v", err)
					_ = d.Nack(false, true)
					continue
				}
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *RabbitConsumer) HandleEmployerUpdated(ctx context.Context, raw []byte) error {
	var evt model.EmployerUpdatedEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	log.Printf("received employer.updated employer_id=%s", evt.EmployerID)
	return c.processor.HandleEmployerUpdated(ctx, evt)
}

func (c *RabbitConsumer) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
