package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	"github.com/study/jobboard/employer-service/internal/model"
)

const (
	exchangeName         = "jobboard"
	routingUserCreated   = "user.created"
	routingEmployerUpd   = "employer.updated"
	queueUserCreated     = "employer.user_created"
)

// Publisher sends domain events via RabbitMQ.
type Publisher interface {
	PublishEmployerUpdated(ctx context.Context, employerID uuid.UUID, companyName string, city *string) error
	Close() error
}

// RabbitPublisher publishes employer.updated.
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

func (p *RabbitPublisher) PublishEmployerUpdated(ctx context.Context, employerID uuid.UUID, companyName string, city *string) error {
	payload := model.EmployerUpdatedEvent{
		Event:       routingEmployerUpd,
		EmployerID:  employerID,
		CompanyName: companyName,
		City:        city,
		OccurredAt:  time.Now().UTC(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, exchangeName, routingEmployerUpd, false, false, amqp.Publishing{
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

// UserCreatedProcessor is implemented by EmployerService.
type UserCreatedProcessor interface {
	HandleUserCreated(ctx context.Context, evt model.UserCreatedEvent) error
}

// RabbitConsumer listens for user.created.
type RabbitConsumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	processor UserCreatedProcessor
}

func NewRabbitConsumer(url string, processor UserCreatedProcessor) (*RabbitConsumer, error) {
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

	_, err = ch.QueueDeclare(queueUserCreated, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(queueUserCreated, routingUserCreated, exchangeName, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	return &RabbitConsumer{conn: conn, ch: ch, processor: processor}, nil
}

func (c *RabbitConsumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume(queueUserCreated, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		log.Printf("rabbit consumer started queue=%s", queueUserCreated)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				if err := c.HandleUserCreated(ctx, d.Body); err != nil {
					log.Printf("user.created handle error: %v", err)
					_ = d.Nack(false, true)
					continue
				}
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *RabbitConsumer) HandleUserCreated(ctx context.Context, raw []byte) error {
	var evt model.UserCreatedEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	log.Printf("received user.created user_id=%s role=%s", evt.UserID, evt.Role)
	return c.processor.HandleUserCreated(ctx, evt)
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
