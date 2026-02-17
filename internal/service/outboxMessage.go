package service

import (
	"context"
	"fmt"
	"time"

	"refina-transaction/config/log"
	"refina-transaction/interface/queue"
	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/model"
	"refina-transaction/internal/utils/data"

	"github.com/rabbitmq/amqp091-go"
)

type OutboxPublisher struct {
	outboxRepo repository.OutboxRepository
	queue      queue.RabbitMQClient
	interval   time.Duration
	batchSize  int
}

func NewOutboxPublisher(
	outboxRepo repository.OutboxRepository,
	rabbitMQ queue.RabbitMQClient,
) *OutboxPublisher {
	return &OutboxPublisher{
		outboxRepo: outboxRepo,
		queue:      rabbitMQ,
		interval:   data.OUTBOX_PUBLISH_INTERVAL,
		batchSize:  data.OUTBOX_PUBLISH_BATCH,
	}
}

// Start begins the outbox publisher worker
func (p *OutboxPublisher) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	log.Log.Info("Outbox publisher started")

	for {
		select {
		case <-ctx.Done():
			log.Log.Info("Outbox publisher stopped")
			return
		case <-ticker.C:
			if err := p.publishPendingMessages(ctx); err != nil {
				log.Log.Errorf("Error publishing outbox messages: %v", err)
			}
		}
	}
}

func (p *OutboxPublisher) publishPendingMessages(ctx context.Context) error {
	messages, err := p.outboxRepo.GetPendingMessages(ctx, p.batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending messages: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	log.Log.Infof("Publishing %d outbox messages", len(messages))

	for _, msg := range messages {
		if err := p.publishMessage(ctx, msg); err != nil {
			log.Log.Errorf("Failed to publish message %d: %v", msg.ID, err)

			// Increment retry count
			if err := p.outboxRepo.IncrementRetries(ctx, msg.ID); err != nil {
				log.Log.Errorf("Failed to increment retries for message %d: %v", msg.ID, err)
			}

			continue
		}

		// Mark as published
		if err := p.outboxRepo.MarkAsPublished(ctx, msg.ID); err != nil {
			log.Log.Errorf("Failed to mark message %d as published: %v", msg.ID, err)
			continue
		}

		log.Log.Infof("Successfully published message %d (event: %s)", msg.ID, msg.EventType)
	}

	return nil
}

func (p *OutboxPublisher) publishMessage(ctx context.Context, msg model.OutboxMessage) error {
	ch, err := p.queue.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		data.OUTBOX_PUBLISH_EXCHANGE,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	message := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         msg.Payload,
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now(),
		MessageId:    fmt.Sprintf("%d", msg.ID),
	}

	return ch.PublishWithContext(
		ctx,
		data.OUTBOX_PUBLISH_EXCHANGE,
		msg.EventType,
		false, // mandatory
		false, // immediate
		message,
	)
}

// StartCleanupJob removes old published messages
func (p *OutboxPublisher) StartCleanupJob(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Log.Info("Outbox cleanup job started")

	for {
		select {
		case <-ctx.Done():
			log.Log.Info("Outbox cleanup job stopped")
			return
		case <-ticker.C:
			if err := p.cleanupOldMessages(ctx); err != nil {
				log.Log.Errorf("Error cleaning up old messages: %v", err)
			}
		}
	}
}

func (p *OutboxPublisher) cleanupOldMessages(ctx context.Context) error {
	// This would need a method in the repository
	// For now, just log
	log.Log.Info("Cleanup job executed")
	return nil
}
