package queue

import (
	"context"
	"fmt"
	"sync"

	"refina-transaction/config/env"
	"refina-transaction/config/log"
	"refina-transaction/internal/utils/data"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient interface {
	GetChannel() (*amqp091.Channel, error)
	Close() error
	Publish(ctx context.Context, routingKey string, body []byte) error
}

type rabbitMQClient struct {
	connection *amqp091.Connection
	mu         sync.RWMutex
}

var (
	instance RabbitMQClient
	once     sync.Once
)

func NewRabbitMQClient(cfg env.RabbitMQ) (RabbitMQClient, error) {
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.RMQUser,
		cfg.RMQPassword,
		cfg.RMQHost,
		cfg.RMQPort,
		cfg.RMQVirtualHost,
	)

	conn, err := amqp091.Dial(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	err = channel.ExchangeDeclare(
		data.OUTBOX_PUBLISH_EXCHANGE,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &rabbitMQClient{
		connection: conn,
	}, nil
}

func GetInstance(cfg env.RabbitMQ) RabbitMQClient {
	once.Do(func() {
		client, err := NewRabbitMQClient(cfg)
		if err != nil {
			log.Log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
		}
		instance = client
	})

	return instance
}

func (r *rabbitMQClient) GetChannel() (*amqp091.Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.connection == nil {
		return nil, fmt.Errorf("RabbitMQ connection is not initialized")
	}

	channel, err := r.connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return channel, nil
}

func (r *rabbitMQClient) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connection != nil {
		if err := r.connection.Close(); err != nil {
			return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
		}
		r.connection = nil
	}

	return nil
}

func (r *rabbitMQClient) Publish(ctx context.Context, routingKey string, body []byte) error {
	channel, err := r.GetChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	message := amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	if err := channel.PublishWithContext(ctx, "refina", routingKey, false, false, message); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
