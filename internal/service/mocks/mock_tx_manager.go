package mocks

import (
	"context"

	"refina-transaction/internal/repository"

	"github.com/stretchr/testify/mock"
)

type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) Begin(ctx context.Context) (repository.Transaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.Transaction), args.Error(1)
}

type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}
