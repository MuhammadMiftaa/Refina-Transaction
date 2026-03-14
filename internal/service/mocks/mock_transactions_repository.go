package mocks

import (
	"context"

	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/model"

	"github.com/stretchr/testify/mock"
)

type MockTransactionsRepository struct {
	mock.Mock
}

func (m *MockTransactionsRepository) GetAllTransactions(ctx context.Context, tx repository.Transaction) ([]model.Transactions, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]model.Transactions), args.Error(1)
}

func (m *MockTransactionsRepository) GetTransactionByID(ctx context.Context, tx repository.Transaction, id string) (model.Transactions, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(model.Transactions), args.Error(1)
}

func (m *MockTransactionsRepository) GetTransactionsByWalletIDs(ctx context.Context, tx repository.Transaction, ids []string) ([]model.Transactions, error) {
	args := m.Called(ctx, tx, ids)
	return args.Get(0).([]model.Transactions), args.Error(1)
}

func (m *MockTransactionsRepository) GetTransactionsByCursor(ctx context.Context, tx repository.Transaction, q repository.CursorQuery) ([]model.Transactions, int64, error) {
	args := m.Called(ctx, tx, q)
	return args.Get(0).([]model.Transactions), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionsRepository) CreateTransaction(ctx context.Context, tx repository.Transaction, transaction model.Transactions) (model.Transactions, error) {
	args := m.Called(ctx, tx, transaction)
	return args.Get(0).(model.Transactions), args.Error(1)
}

func (m *MockTransactionsRepository) UpdateTransaction(ctx context.Context, tx repository.Transaction, transaction model.Transactions) (model.Transactions, error) {
	args := m.Called(ctx, tx, transaction)
	return args.Get(0).(model.Transactions), args.Error(1)
}

func (m *MockTransactionsRepository) DeleteTransaction(ctx context.Context, tx repository.Transaction, transaction model.Transactions) (model.Transactions, error) {
	args := m.Called(ctx, tx, transaction)
	return args.Get(0).(model.Transactions), args.Error(1)
}
