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

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.publishPendingMessages(ctx); err != nil {
				log.Error(data.LogOutboxPublishPendingFailed, map[string]any{"service": data.OutboxService, "error": err.Error()})
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

	for _, msg := range messages {
		if err := p.publishMessage(ctx, msg); err != nil {
			log.Error(data.LogOutboxMessagePublishFailed, map[string]any{
				"service":     data.OutboxService,
				"document_id": msg.ID,
				"event_type":  msg.EventType,
				"error":       err.Error(),
			})

			if msg.Retries >= msg.MaxRetries-1 {
				log.Error(data.LogOutboxMessageMaxRetriesExceeded, map[string]any{
					"service":     data.OutboxService,
					"document_id": msg.ID,
					"event_type":  msg.EventType,
					"retries":     msg.Retries,
				})
			}

			// Increment retry count
			if err := p.outboxRepo.IncrementRetries(ctx, msg.ID); err != nil {
				log.Error(data.LogOutboxIncrementRetriesFailed, map[string]any{
					"service":     data.OutboxService,
					"document_id": msg.ID,
					"event_type":  msg.EventType,
					"error":       err.Error(),
				})
			}

			continue
		}

		// Mark as published
		if err := p.outboxRepo.MarkAsPublished(ctx, msg.ID); err != nil {
			log.Error(data.LogOutboxMarkPublishedFailed, map[string]any{
				"service":     data.OutboxService,
				"document_id": msg.ID,
				"event_type":  msg.EventType,
				"error":       err.Error(),
			})
			continue
		}

		log.Info(data.LogOutboxMessagePublished, map[string]any{
			"service":     data.OutboxService,
			"document_id": msg.ID,
			"event_type":  msg.EventType,
		})
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.cleanupOldMessages(ctx); err != nil {
				log.Error(data.LogOutboxCleanupFailed, map[string]any{"service": data.OutboxService, "error": err.Error()})
			}
		}
	}
}

func (p *OutboxPublisher) cleanupOldMessages(ctx context.Context) error {
	// This would need a method in the repository
	return nil
}
