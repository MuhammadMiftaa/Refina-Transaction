package mocks

import (
	"context"

	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/model"

	"github.com/stretchr/testify/mock"
)

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) Create(ctx context.Context, tx repository.Transaction, outbox *model.OutboxMessage) error {
	args := m.Called(ctx, tx, outbox)
	return args.Error(0)
}

func (m *MockOutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]model.OutboxMessage, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]model.OutboxMessage), args.Error(1)
}

func (m *MockOutboxRepository) MarkAsPublished(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOutboxRepository) IncrementRetries(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
