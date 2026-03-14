package mocks

import (
	"context"

	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/model"
	"refina-transaction/internal/types/view"

	"github.com/stretchr/testify/mock"
)

type MockCategoriesRepository struct {
	mock.Mock
}

func (m *MockCategoriesRepository) GetAllCategories(ctx context.Context, tx repository.Transaction) ([]model.Categories, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]model.Categories), args.Error(1)
}

func (m *MockCategoriesRepository) GetCategoryByID(ctx context.Context, tx repository.Transaction, id string) (model.Categories, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(model.Categories), args.Error(1)
}

func (m *MockCategoriesRepository) GetCategoriesByType(ctx context.Context, tx repository.Transaction, typeCategory string) ([]view.ViewCategoriesGroupByType, error) {
	args := m.Called(ctx, tx, typeCategory)
	return args.Get(0).([]view.ViewCategoriesGroupByType), args.Error(1)
}

func (m *MockCategoriesRepository) CreateCategory(ctx context.Context, tx repository.Transaction, category model.Categories) (model.Categories, error) {
	args := m.Called(ctx, tx, category)
	return args.Get(0).(model.Categories), args.Error(1)
}

func (m *MockCategoriesRepository) UpdateCategory(ctx context.Context, tx repository.Transaction, category model.Categories) (model.Categories, error) {
	args := m.Called(ctx, tx, category)
	return args.Get(0).(model.Categories), args.Error(1)
}

func (m *MockCategoriesRepository) DeleteCategory(ctx context.Context, tx repository.Transaction, category model.Categories) (model.Categories, error) {
	args := m.Called(ctx, tx, category)
	return args.Get(0).(model.Categories), args.Error(1)
}
