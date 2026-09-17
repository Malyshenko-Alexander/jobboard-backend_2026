package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
)

const (
	exchangeName     = "jobboard"
	routingUserCreated = "user.created"
)

// UserCreated is published after successful registration.
type UserCreated struct {
	Event      string    `json:"event"`
	UserID     uuid.UUID `json:"user_id"`
	Role       string    `json:"role"`
	Email      string    `json:"email"`
	OccurredAt time.Time `json:"occurred_at"`
}

// Publisher sends domain events to other microservices via RabbitMQ.
type Publisher interface {
	PublishUserCreated(ctx context.Context, userID uuid.UUID, role, email string) error
	Close() error
}

// RabbitPublisher publishes to a topic exchange.
type RabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitPublisher connects to RabbitMQ and declares the shared exchange.
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

func (p *RabbitPublisher) PublishUserCreated(ctx context.Context, userID uuid.UUID, role, email string) error {
	payload := UserCreated{
		Event:      routingUserCreated,
		UserID:     userID,
		Role:       role,
		Email:      email,
		OccurredAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, exchangeName, routingUserCreated, false, false, amqp.Publishing{
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
