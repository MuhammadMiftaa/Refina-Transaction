package repository

import (
	"context"
	"errors"

	"refina-transaction/internal/types/model"

	"gorm.io/gorm"
)

type TransactionsRepository interface {
	GetAllTransactions(ctx context.Context, tx Transaction) ([]model.Transactions, error)
	GetTransactionByID(ctx context.Context, tx Transaction, id string) (model.Transactions, error)
	GetTransactionsByUserID(ctx context.Context, tx Transaction, id string) ([]model.Transactions, error)
	CreateTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error)
	UpdateTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error)
	DeleteTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error)
}

type transactionsRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionsRepository {
	return &transactionsRepository{db}
}

func (transaction_repo *transactionsRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return transaction_repo.db.WithContext(ctx), nil
}

func (transaction_repo *transactionsRepository) GetAllTransactions(ctx context.Context, tx Transaction) ([]model.Transactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var transactions []model.Transactions
	err = db.Preload("Category").Order("transaction_date DESC").Find(&transactions).Error
	if err != nil {
		return nil, errors.New("user transactions not found")
	}
	return transactions, nil
}

func (trasaction_repo *transactionsRepository) GetTransactionByID(ctx context.Context, tx Transaction, id string) (model.Transactions, error) {
	db, err := trasaction_repo.getDB(ctx, tx)
	if err != nil {
		return model.Transactions{}, err
	}

	var transaction model.Transactions
	err = db.Preload("Category").Where("id = ?", id).First(&transaction).Error
	if err != nil {
		return model.Transactions{}, errors.New("transaction not found")
	}

	return transaction, nil
}

func (transaction_repo *transactionsRepository) GetTransactionsByUserID(ctx context.Context, tx Transaction, id string) ([]model.Transactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var userTransactions []model.Transactions
	err = db.Preload("Category").Where("user_id = ?", id).Order("transaction_date DESC").Find(&userTransactions).Error
	if err != nil {
		return nil, errors.New("user transactions not found")
	}
	return userTransactions, nil
}

func (transaction_repo *transactionsRepository) CreateTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return model.Transactions{}, err
	}

	if err := db.Create(&transaction).Error; err != nil {
		return model.Transactions{}, err
	}

	return transaction, nil
}

func (transaction_repo *transactionsRepository) UpdateTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return model.Transactions{}, err
	}

	if err := db.Omit("Wallet", "Category").Save(&transaction).Error; err != nil {
		return model.Transactions{}, err
	}

	return transaction, nil
}

func (transaction_repo *transactionsRepository) DeleteTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return model.Transactions{}, err
	}

	if err := db.Delete(&transaction).Error; err != nil {
		return model.Transactions{}, err
	}
	return transaction, nil
}
