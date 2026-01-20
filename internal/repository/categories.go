package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"refina-transaction/internal/types/model"
	"refina-transaction/internal/types/view"

	"gorm.io/gorm"
)

type CategoriesRepository interface {
	GetAllCategories(ctx context.Context, tx Transaction) ([]model.Categories, error)
	GetCategoryByID(ctx context.Context, tx Transaction, id string) (model.Categories, error)
	GetCategoriesByType(ctx context.Context, tx Transaction, typeCategory string) ([]view.ViewCategoriesGroupByType, error)
	CreateCategory(ctx context.Context, tx Transaction, category model.Categories) (model.Categories, error)
	UpdateCategory(ctx context.Context, tx Transaction, category model.Categories) (model.Categories, error)
	DeleteCategory(ctx context.Context, tx Transaction, category model.Categories) (model.Categories, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoriesRepository {
	return &categoryRepository{db}
}

// Helper untuk mendapatkan DB instance (transaksi atau biasa)
func (category_repo *categoryRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx) // Type assertion ke GORM transaction
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return category_repo.db.WithContext(ctx), nil
}

func (category_repo *categoryRepository) GetAllCategories(ctx context.Context, tx Transaction) ([]model.Categories, error) {
	db, err := category_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var categories []model.Categories
	if err := db.Preload("Parent").Preload("Children").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (category_repo *categoryRepository) GetCategoryByID(ctx context.Context, tx Transaction, id string) (model.Categories, error) {
	db, err := category_repo.getDB(ctx, tx)
	if err != nil {
		return model.Categories{}, err
	}

	var category model.Categories
	if err := db.Preload("Parent").Preload("Children").First(&category, "id = ?", id).Error; err != nil {
		return model.Categories{}, err
	}

	return category, nil
}

func (category_repo *categoryRepository) GetCategoriesByType(ctx context.Context, tx Transaction, typeCategory string) ([]view.ViewCategoriesGroupByType, error) {
	db, err := category_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var rawResults []struct {
		GroupName string
		Type      string
		Category  []byte
	}
	err = db.Raw(`SELECT * FROM view_category_group_by_type WHERE type = $1`, typeCategory).Scan(&rawResults).Error
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		return nil, errors.New("category group by type not found")
	}

	var results []view.ViewCategoriesGroupByType

	for _, row := range rawResults {
		var categories []view.ViewCategoriesGroupByTypeDetail

		err := json.Unmarshal(row.Category, &categories)
		if err != nil {
			return nil, fmt.Errorf("gagal decode JSON categories (type: %s): %w", row.Type, err)
		}

		results = append(results, view.ViewCategoriesGroupByType{
			GroupName: row.GroupName,
			Type:      row.Type,
			Category:  categories,
		})
	}

	return results, nil
}

func (category_repo *categoryRepository) CreateCategory(ctx context.Context, tx Transaction, category model.Categories) (model.Categories, error) {
	db, err := category_repo.getDB(ctx, tx)
	if err != nil {
		return model.Categories{}, err
	}

	if err := db.Create(&category).Error; err != nil {
		return model.Categories{}, err
	}

	return category, nil
}

func (category_repo *categoryRepository) UpdateCategory(ctx context.Context, tx Transaction, category model.Categories) (model.Categories, error) {
	db, err := category_repo.getDB(ctx, tx)
	if err != nil {
		return model.Categories{}, err
	}

	if err := db.Save(&category).Error; err != nil {
		return model.Categories{}, err
	}

	return category, nil
}

func (category_repo *categoryRepository) DeleteCategory(ctx context.Context, tx Transaction, category model.Categories) (model.Categories, error) {
	db, err := category_repo.getDB(ctx, tx)
	if err != nil {
		return model.Categories{}, err
	}

	if err := db.Delete(&category).Error; err != nil {
		return model.Categories{}, err
	}

	return category, nil
}
