package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"refina-transaction/config/log"
	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"
	"refina-transaction/internal/utils"
	"refina-transaction/internal/utils/data"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
)

type transactionServer struct {
	tpb.UnimplementedTransactionServiceServer
	txManager              repository.TxManager
	transactionsRepository repository.TransactionsRepository
	categoryRepository     repository.CategoriesRepository
	outboxRepository       repository.OutboxRepository
}

func (s *transactionServer) GetTransactions(req *tpb.GetTransactionOptions, stream tpb.TransactionService_GetUserTransactionsServer) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	transactions, err := s.transactionsRepository.GetAllTransactions(timeout, nil)
	if err != nil {
		log.Error(data.LogGetTransactionsFailed, map[string]any{
			"service": data.GRPCServerService,
			"error":   err.Error(),
		})
		return fmt.Errorf("get transactions: %w", err)
	}

	for _, transaction := range transactions {
		if err := stream.Send(&tpb.Transaction{
			Id:              transaction.ID.String(),
			WalletId:        transaction.WalletID.String(),
			Amount:          transaction.Amount,
			CategoryId:      transaction.CategoryID.String(),
			CategoryName:    transaction.Category.Name,
			CategoryType:    string(transaction.Category.Type),
			TransactionDate: transaction.TransactionDate.Format(time.RFC3339),
			Description:     transaction.Description,
			CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
		}); err != nil {
			log.Error(data.LogStreamSendFailed, map[string]any{
				"service":        data.GRPCServerService,
				"transaction_id": transaction.ID.String(),
				"error":          err.Error(),
			})
			return fmt.Errorf("stream send [transaction_id=%s]: %w", transaction.ID.String(), err)
		}
	}

	return nil
}

func (s *transactionServer) GetUserTransactions(req *tpb.Wallets, stream tpb.TransactionService_GetUserTransactionsServer) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	transactions, err := s.transactionsRepository.GetTransactionsByWalletIDs(timeout, nil, req.GetWalletId())
	if err != nil {
		log.Error(data.LogGetTransactionsFailed, map[string]any{
			"service": data.GRPCServerService,
			"error":   err.Error(),
		})
		return fmt.Errorf("get user transactions: %w", err)
	}

	for _, transaction := range transactions {
		if err := stream.Send(&tpb.Transaction{
			Id:              transaction.ID.String(),
			WalletId:        transaction.WalletID.String(),
			Amount:          transaction.Amount,
			CategoryId:      transaction.CategoryID.String(),
			CategoryName:    transaction.Category.Name,
			CategoryType:    string(transaction.Category.Type),
			TransactionDate: transaction.TransactionDate.Format(time.RFC3339),
			Description:     transaction.Description,
			CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
		}); err != nil {
			log.Error(data.LogStreamSendFailed, map[string]any{
				"service":        data.GRPCServerService,
				"transaction_id": transaction.ID.String(),
				"error":          err.Error(),
			})
			return fmt.Errorf("stream send [transaction_id=%s]: %w", transaction.ID.String(), err)
		}
	}

	return nil
}

func (s *transactionServer) CreateTransaction(ctx context.Context, req *tpb.NewTransaction) (*tpb.Transaction, error) {
	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create transaction: begin transaction: %w", err)
	}

	defer tx.Rollback()

	walletID, _ := utils.ParseUUID(req.GetWalletId())
	categoryID, _ := utils.ParseUUID(req.GetCategoryId())
	transactionDate, err := time.Parse(time.RFC3339, req.GetTransactionDate())
	if err != nil {
		return nil, fmt.Errorf("create transaction: parse date: %w", err)
	}

	category, err := s.categoryRepository.GetCategoryByID(ctx, nil, req.GetCategoryId())
	if err != nil {
		return nil, fmt.Errorf("category not found [id=%s]: %w", req.GetCategoryId(), err)
	}

	if req.GetAmount() <= 0 {
		return nil, fmt.Errorf("invalid transaction amount [amount=%v]", req.GetAmount())
	}

	transaction, err := s.transactionsRepository.CreateTransaction(ctx, nil, model.Transactions{
		WalletID:        walletID,
		Amount:          req.GetAmount(),
		CategoryID:      categoryID,
		TransactionDate: transactionDate,
		Description:     req.GetDescription(),
		Category:        category,
	})
	if err != nil {
		return nil, fmt.Errorf("create transaction: insert to db: %w", err)
	}

	transactionResponse := utils.ConvertToResponseType(transaction).(dto.TransactionsResponse)
	payload, err := json.Marshal(transactionResponse)
	if err != nil {
		return nil, fmt.Errorf("create transaction: marshal response: %w", err)
	}

	outboxMsg := &model.OutboxMessage{
		AggregateID: transactionResponse.ID,
		EventType:   data.OUTBOX_EVENT_TRANSACTION_CREATED,
		Payload:     payload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}

	if err := s.outboxRepository.Create(ctx, tx, outboxMsg); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create transaction: commit: %w", err)
	}

	log.Info(data.LogTransactionCreated, map[string]any{
		"service":        data.GRPCServerService,
		"transaction_id": transaction.ID.String(),
		"wallet_id":      transaction.WalletID.String(),
	})

	return &tpb.Transaction{
		Id:              transaction.ID.String(),
		WalletId:        transaction.WalletID.String(),
		Amount:          transaction.Amount,
		CategoryId:      transaction.CategoryID.String(),
		CategoryName:    transaction.Category.Name,
		CategoryType:    string(transaction.Category.Type),
		TransactionDate: transaction.TransactionDate.Format(time.RFC3339),
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *transactionServer) DeleteTransaction(ctx context.Context, req *tpb.TransactionID) (*tpb.Transaction, error) {
	transactionID := req.GetId()

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete transaction: begin transaction: %w", err)
	}

	defer tx.Rollback()

	transaction, err := s.transactionsRepository.GetTransactionByID(ctx, tx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found [id=%s]: %w", transactionID, err)
	}

	transactionDeleted, err := s.transactionsRepository.DeleteTransaction(ctx, tx, transaction)
	if err != nil {
		return nil, fmt.Errorf("delete transaction [id=%s]: delete from db: %w", transactionID, err)
	}

	transactionResponse := utils.ConvertToResponseType(transactionDeleted).(dto.TransactionsResponse)
	payload, err := json.Marshal(transactionResponse)
	if err != nil {
		return nil, fmt.Errorf("delete transaction: marshal response: %w", err)
	}

	outboxMsg := &model.OutboxMessage{
		AggregateID: transactionResponse.ID,
		EventType:   data.OUTBOX_EVENT_TRANSACTION_DELETED,
		Payload:     payload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}

	if err := s.outboxRepository.Create(ctx, tx, outboxMsg); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("delete transaction: commit: %w", err)
	}

	log.Info(data.LogTransactionDeleted, map[string]any{
		"service":        data.GRPCServerService,
		"transaction_id": transactionID,
	})

	return &tpb.Transaction{
		Id:              transaction.ID.String(),
		WalletId:        transaction.WalletID.String(),
		Amount:          transaction.Amount,
		CategoryId:      transaction.CategoryID.String(),
		CategoryName:    transaction.Category.Name,
		CategoryType:    string(transaction.Category.Type),
		TransactionDate: transaction.TransactionDate.Format(time.RFC3339),
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
	}, nil
}
