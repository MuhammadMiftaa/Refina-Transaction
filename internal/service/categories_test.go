package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"refina-transaction/internal/service/mocks"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"
	"refina-transaction/internal/types/view"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// Test Dependency Container
// ─────────────────────────────────────────────

type categoryTestDeps struct {
	txManager    *mocks.MockTxManager
	categoryRepo *mocks.MockCategoriesRepository
	tx           *mocks.MockTransaction
}

func newCategoryTestDeps() *categoryTestDeps {
	return &categoryTestDeps{
		txManager:    new(mocks.MockTxManager),
		categoryRepo: new(mocks.MockCategoriesRepository),
		tx:           new(mocks.MockTransaction),
	}
}

func (d *categoryTestDeps) service() CategoriesService {
	return NewCategoriesService(d.txManager, d.categoryRepo)
}

func (d *categoryTestDeps) assertAll(t *testing.T) {
	t.Helper()
	d.txManager.AssertExpectations(t)
	d.categoryRepo.AssertExpectations(t)
	d.tx.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Fixed UUIDs & Timestamps
// ─────────────────────────────────────────────

var (
	catParentID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	catChildID  = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	catFixTime  = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
)

// ─────────────────────────────────────────────
// Sample Data Factories
// ─────────────────────────────────────────────

func sampleParentCategory() model.Categories {
	return model.Categories{
		Base:     model.Base{ID: catParentID, CreatedAt: catFixTime, UpdatedAt: catFixTime},
		ParentID: nil,
		Name:     "Transportasi",
		Type:     model.Expense,
	}
}

func sampleChildCategory() model.Categories {
	parent := sampleParentCategory()
	return model.Categories{
		Base:     model.Base{ID: catChildID, CreatedAt: catFixTime, UpdatedAt: catFixTime},
		ParentID: &catParentID,
		Name:     "Parkir",
		Type:     model.Expense,
		Parent:   &parent,
	}
}

func sampleViewCategories() []view.ViewCategoriesGroupByType {
	return []view.ViewCategoriesGroupByType{
		{
			GroupName: "Transportasi",
			Type:      "expense",
			Category: []view.ViewCategoriesGroupByTypeDetail{
				{ID: catChildID.String(), Name: "Parkir"},
			},
		},
	}
}

// =====================================================================
// GetAllCategories
// =====================================================================

func TestGetAllCategories_Success(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	parent := sampleParentCategory()
	child := sampleChildCategory()
	categories := []model.Categories{parent, child}

	d.categoryRepo.On("GetAllCategories", mock.Anything, nil).Return(categories, nil)

	result, err := svc.GetAllCategories(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Parent becomes the group, child becomes the category item
	assert.Len(t, result, 1)
	assert.Equal(t, "Transportasi", result[0].GroupName)
	assert.Len(t, result[0].Category, 1)
	assert.Equal(t, catChildID.String(), result[0].Category[0].ID)
	d.assertAll(t)
}

func TestGetAllCategories_EmptyList(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetAllCategories", mock.Anything, nil).Return([]model.Categories{}, nil)

	result, err := svc.GetAllCategories(context.Background())

	assert.NoError(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

func TestGetAllCategories_RepositoryError(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetAllCategories", mock.Anything, nil).
		Return([]model.Categories{}, errors.New("db error"))

	result, err := svc.GetAllCategories(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

// =====================================================================
// GetCategoryByID
// =====================================================================

func TestGetCategoryByID_SuccessChild(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	child := sampleChildCategory()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(child, nil)

	result, err := svc.GetCategoryByID(context.Background(), catChildID.String())

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	assert.Equal(t, dto.CategoryType(model.Expense), result.Type)
	assert.Len(t, result.Category, 1)
	assert.Equal(t, catChildID.String(), result.Category[0].ID)
	d.assertAll(t)
}

func TestGetCategoryByID_SuccessParent(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	parent := sampleParentCategory()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catParentID.String()).Return(parent, nil)

	result, err := svc.GetCategoryByID(context.Background(), catParentID.String())

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	assert.Nil(t, result.Category) // parent has no category items
	d.assertAll(t)
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, "nonexistent").
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.GetCategoryByID(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

// =====================================================================
// GetCategoriesByType
// =====================================================================

func TestGetCategoriesByType_Success(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	views := sampleViewCategories()
	d.categoryRepo.On("GetCategoriesByType", mock.Anything, nil, "expense").Return(views, nil)

	result, err := svc.GetCategoriesByType(context.Background(), "expense")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Transportasi", result[0].GroupName)
	d.assertAll(t)
}

func TestGetCategoriesByType_RepositoryError(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetCategoriesByType", mock.Anything, nil, "expense").
		Return([]view.ViewCategoriesGroupByType{}, errors.New("db error"))

	result, err := svc.GetCategoriesByType(context.Background(), "expense")

	assert.Error(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

func TestGetCategoriesByType_EmptyResult(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetCategoriesByType", mock.Anything, nil, "unknown_type").
		Return([]view.ViewCategoriesGroupByType{}, nil)

	result, err := svc.GetCategoriesByType(context.Background(), "unknown_type")

	assert.NoError(t, err)
	assert.Empty(t, result)
	d.assertAll(t)
}

// =====================================================================
// CreateCategory
// =====================================================================

func TestCreateCategory_SuccessWithParent(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	parent := sampleParentCategory()
	newChild := sampleChildCategory()
	req := dto.CategoriesRequest{
		ParentID: catParentID.String(),
		Name:     "Parkir",
	}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catParentID.String()).Return(parent, nil)
	d.categoryRepo.On("CreateCategory", mock.Anything, nil, mock.MatchedBy(func(c model.Categories) bool {
		return c.Name == "Parkir" && c.ParentID != nil && *c.ParentID == catParentID
	})).Return(newChild, nil)

	result, err := svc.CreateCategory(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	assert.Equal(t, dto.CategoryType(model.Expense), result.Type)
	assert.Len(t, result.Category, 1)
	d.assertAll(t)
}

func TestCreateCategory_SuccessWithoutParent(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	newParent := sampleParentCategory()
	req := dto.CategoriesRequest{
		Name: "Transportasi",
		Type: dto.Expense,
	}

	d.categoryRepo.On("CreateCategory", mock.Anything, nil, mock.MatchedBy(func(c model.Categories) bool {
		return c.Name == "Transportasi" && c.ParentID == nil
	})).Return(newParent, nil)

	result, err := svc.CreateCategory(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	assert.Equal(t, dto.Expense, result.Type)
	d.assertAll(t)
}

func TestCreateCategory_ParentNotFound(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	req := dto.CategoriesRequest{
		ParentID: "non-existent-id",
		Name:     "Parkir",
	}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, "non-existent-id").
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.CreateCategory(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

func TestCreateCategory_RepositoryCreateError(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	req := dto.CategoriesRequest{
		Name: "Transportasi",
		Type: dto.Expense,
	}

	d.categoryRepo.On("CreateCategory", mock.Anything, nil, mock.Anything).
		Return(model.Categories{}, errors.New("db error"))

	result, err := svc.CreateCategory(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

// =====================================================================
// UpdateCategory
// =====================================================================

func TestUpdateCategory_SuccessChangeName(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	existing := sampleChildCategory()
	updated := existing
	updated.Name = "Tol & Parkir"

	req := dto.CategoriesRequest{Name: "Tol & Parkir"}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(existing, nil)
	d.categoryRepo.On("UpdateCategory", mock.Anything, nil, mock.MatchedBy(func(c model.Categories) bool {
		return c.Name == "Tol & Parkir"
	})).Return(updated, nil)

	result, err := svc.UpdateCategory(context.Background(), catChildID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	assert.Len(t, result.Category, 1)
	assert.Equal(t, "Tol & Parkir", result.Category[0].Name)
	d.assertAll(t)
}

func TestUpdateCategory_SuccessChangeParent(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	newParentID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	newParent := model.Categories{
		Base: model.Base{ID: newParentID},
		Name: "Belanja",
		Type: model.Expense,
	}

	existing := sampleChildCategory()
	updatedChild := existing
	updatedChild.ParentID = &newParentID
	updatedChild.Parent = &newParent

	req := dto.CategoriesRequest{ParentID: newParentID.String()}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(existing, nil)
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, newParentID.String()).Return(newParent, nil)
	d.categoryRepo.On("UpdateCategory", mock.Anything, nil, mock.Anything).Return(updatedChild, nil)

	result, err := svc.UpdateCategory(context.Background(), catChildID.String(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.GroupName)
	d.assertAll(t)
}

func TestUpdateCategory_NotFound(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, "bad-id").
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.UpdateCategory(context.Background(), "bad-id", dto.CategoriesRequest{Name: "X"})

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

func TestUpdateCategory_NewParentNotFound(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	existing := sampleChildCategory()
	req := dto.CategoriesRequest{ParentID: "nonexistent-parent"}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(existing, nil)
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, "nonexistent-parent").
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.UpdateCategory(context.Background(), catChildID.String(), req)

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

func TestUpdateCategory_InvalidNewParentUUID(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	existing := sampleChildCategory()
	fakeParent := model.Categories{Base: model.Base{ID: uuid.New()}, Name: "X", Type: model.Expense}
	req := dto.CategoriesRequest{ParentID: "not-a-uuid"}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(existing, nil)
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, "not-a-uuid").Return(fakeParent, nil)

	result, err := svc.UpdateCategory(context.Background(), catChildID.String(), req)

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

func TestUpdateCategory_RepositoryUpdateError(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	existing := sampleChildCategory()
	req := dto.CategoriesRequest{Name: "New Name"}

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(existing, nil)
	d.categoryRepo.On("UpdateCategory", mock.Anything, nil, mock.Anything).
		Return(model.Categories{}, errors.New("db error"))

	result, err := svc.UpdateCategory(context.Background(), catChildID.String(), req)

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

// =====================================================================
// DeleteCategory
// =====================================================================

func TestDeleteCategory_SuccessChild(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	child := sampleChildCategory()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(child, nil)
	d.categoryRepo.On("DeleteCategory", mock.Anything, nil, child).Return(child, nil)

	result, err := svc.DeleteCategory(context.Background(), catChildID.String())

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	assert.Len(t, result.Category, 1)
	d.assertAll(t)
}

func TestDeleteCategory_SuccessParent(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	parent := sampleParentCategory()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catParentID.String()).Return(parent, nil)
	d.categoryRepo.On("DeleteCategory", mock.Anything, nil, parent).Return(parent, nil)

	result, err := svc.DeleteCategory(context.Background(), catParentID.String())

	assert.NoError(t, err)
	assert.Equal(t, "Transportasi", result.GroupName)
	d.assertAll(t)
}

func TestDeleteCategory_NotFound(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, "bad-id").
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.DeleteCategory(context.Background(), "bad-id")

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}

func TestDeleteCategory_RepositoryDeleteError(t *testing.T) {
	d := newCategoryTestDeps()
	svc := d.service()

	child := sampleChildCategory()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, nil, catChildID.String()).Return(child, nil)
	d.categoryRepo.On("DeleteCategory", mock.Anything, nil, child).
		Return(model.Categories{}, errors.New("db error"))

	result, err := svc.DeleteCategory(context.Background(), catChildID.String())

	assert.Error(t, err)
	assert.Empty(t, result.GroupName)
	d.assertAll(t)
}
