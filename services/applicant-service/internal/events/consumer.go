package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/study/jobboard/applicant-service/internal/model"
)

const (
	exchangeName       = "jobboard"
	routingUserCreated = "user.created"
	queueUserCreated   = "applicant.user_created"
)

// UserCreatedProcessor is implemented by ApplicantService.
type UserCreatedProcessor interface {
	HandleUserCreated(ctx context.Context, evt model.UserCreatedEvent) error
}

// RabbitConsumer listens for user.created events.
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

// Start begins consuming in a background goroutine until ctx is cancelled.
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

// HandleUserCreated processes payload (also used by HTTP debug hook).
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
