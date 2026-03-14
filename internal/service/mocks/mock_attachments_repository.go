package mocks

import (
	"context"

	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/model"

	"github.com/stretchr/testify/mock"
)

type MockAttachmentsRepository struct {
	mock.Mock
}

func (m *MockAttachmentsRepository) GetAllAttachments(ctx context.Context, tx repository.Transaction) ([]model.Attachments, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]model.Attachments), args.Error(1)
}

func (m *MockAttachmentsRepository) GetAttachmentByID(ctx context.Context, tx repository.Transaction, id string) (model.Attachments, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(model.Attachments), args.Error(1)
}

func (m *MockAttachmentsRepository) GetAttachmentsByTransactionID(ctx context.Context, tx repository.Transaction, transactionID string) ([]model.Attachments, error) {
	args := m.Called(ctx, tx, transactionID)
	return args.Get(0).([]model.Attachments), args.Error(1)
}

func (m *MockAttachmentsRepository) CreateAttachment(ctx context.Context, tx repository.Transaction, attachment model.Attachments) (model.Attachments, error) {
	args := m.Called(ctx, tx, attachment)
	return args.Get(0).(model.Attachments), args.Error(1)
}

func (m *MockAttachmentsRepository) UpdateAttachment(ctx context.Context, tx repository.Transaction, attachment model.Attachments) (model.Attachments, error) {
	args := m.Called(ctx, tx, attachment)
	return args.Get(0).(model.Attachments), args.Error(1)
}

func (m *MockAttachmentsRepository) DeleteAttachment(ctx context.Context, tx repository.Transaction, attachment model.Attachments) (model.Attachments, error) {
	args := m.Called(ctx, tx, attachment)
	return args.Get(0).(model.Attachments), args.Error(1)
}
