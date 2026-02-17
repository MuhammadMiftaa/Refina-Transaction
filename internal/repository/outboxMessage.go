package repository

import (
	"context"
	"errors"

	"refina-transaction/internal/types/model"

	"gorm.io/gorm"
)

type OutboxRepository interface {
	Create(ctx context.Context, tx Transaction, outbox *model.OutboxMessage) error
	GetPendingMessages(ctx context.Context, limit int) ([]model.OutboxMessage, error)
	MarkAsPublished(ctx context.Context, id uint) error
	IncrementRetries(ctx context.Context, id uint) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *outboxRepository) Create(ctx context.Context, tx Transaction, outbox *model.OutboxMessage) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	return db.Create(outbox).Error
}

func (r *outboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]model.OutboxMessage, error) {
	var messages []model.OutboxMessage

	err := r.db.WithContext(ctx).
		Where("published = ?", false).
		Where("retries < max_retries").
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

func (r *outboxRepository) MarkAsPublished(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"published":    true,
			"published_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *outboxRepository) IncrementRetries(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxMessage{}).
		Where("id = ?", id).
		Update("retries", gorm.Expr("retries + 1")).Error
}
