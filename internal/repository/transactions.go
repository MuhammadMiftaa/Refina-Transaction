package repository

import (
	"context"
	"errors"

	"refina-transaction/internal/types/model"
	"refina-transaction/internal/types/view"

	"gorm.io/gorm"
)

type TransactionsRepository interface {
	GetAllTransactions(ctx context.Context, tx Transaction) ([]view.ViewUserTransactions, error)
	GetTransactionByID(ctx context.Context, tx Transaction, id string) (model.Transactions, error)
	GetTransactionByIDJoin(ctx context.Context, tx Transaction, id string) (view.ViewUserTransactions, error)
	GetTransactionsByUserID(ctx context.Context, tx Transaction, id string) ([]view.ViewUserTransactions, error)
	CreateTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error)
	UpdateTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error)
	DeleteTransaction(ctx context.Context, tx Transaction, transaction model.Transactions) (model.Transactions, error)
	GetUserSummary(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserSummaries, error)
	GetUserMonthlySummary(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserMonthlySummaries, error)
	GetUserMostExpenses(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserMostExpenses, error)
	GetUserWalletDailySummary(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserWalletDailySummaries, error)
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

func (transaction_repo *transactionsRepository) GetAllTransactions(ctx context.Context, tx Transaction) ([]view.ViewUserTransactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var transactions []view.ViewUserTransactions
	err = db.Table("view_user_transactions").Order("transaction_date DESC").Find(&transactions).Error
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
	err = db.Preload("Category").Preload("Wallet").Where("id = ?", id).First(&transaction).Error
	if err != nil {
		return model.Transactions{}, errors.New("transaction not found")
	}

	return transaction, nil
}

func (transaction_repo *transactionsRepository) GetTransactionByIDJoin(ctx context.Context, tx Transaction, id string) (view.ViewUserTransactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return view.ViewUserTransactions{}, err
	}

	var transaction view.ViewUserTransactions
	err = db.Table("view_user_transactions").Where("id = ?", id).First(&transaction).Error
	if err != nil {
		return view.ViewUserTransactions{}, errors.New("transaction not found")
	}

	return transaction, nil
}

func (transaction_repo *transactionsRepository) GetTransactionsByUserID(ctx context.Context, tx Transaction, id string) ([]view.ViewUserTransactions, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var userTransactions []view.ViewUserTransactions
	err = db.Table("view_user_transactions").Where("user_id = ?", id).Order("transaction_date DESC").Find(&userTransactions).Error
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

func (transaction_repo *transactionsRepository) GetUserSummary(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserSummaries, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var summaries []view.MVUserSummaries
	query := db.Table("view_user_summaries")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err = query.Find(&summaries).Error
	if err != nil {
		return nil, errors.New("user summaries not found")
	}

	return summaries, nil
}

func (transaction_repo *transactionsRepository) GetUserMonthlySummary(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserMonthlySummaries, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var summaries []view.MVUserMonthlySummaries
	query := db.Table("view_user_monthly_summaries")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err = query.Find(&summaries).Error
	if err != nil {
		return nil, errors.New("user monthly summaries not found")
	}

	return summaries, nil
}

func (transaction_repo *transactionsRepository) GetUserMostExpenses(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserMostExpenses, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var expenses []view.MVUserMostExpenses
	query := db.Table("view_user_most_expenses")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err = query.Find(&expenses).Error
	if err != nil {
		return nil, errors.New("user most expenses not found")
	}

	return expenses, nil
}

func (transaction_repo *transactionsRepository) GetUserWalletDailySummary(ctx context.Context, tx Transaction, userID *string) ([]view.MVUserWalletDailySummaries, error) {
	db, err := transaction_repo.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var summaries []view.MVUserWalletDailySummaries
	query := db.Table("view_user_wallet_daily_summaries")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err = query.Find(&summaries).Error
	if err != nil {
		return nil, errors.New("user wallet daily summaries not found")
	}

	return summaries, nil
}
