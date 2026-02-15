package server

import (
	"context"
	"time"

	"refina-transaction/internal/repository"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
)

type transactionServer struct {
	tpb.UnimplementedTransactionServiceServer
	transactionsRepository repository.TransactionsRepository
}

func (s *transactionServer) GetTransactions(req *tpb.GetTransactionOptions, stream tpb.TransactionService_GetUserTransactionsServer) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	transactions, err := s.transactionsRepository.GetAllTransactions(timeout, nil)
	if err != nil {
		return err
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
			return err
		}
	}

	return nil
}


func (s *transactionServer) GetUserTransactions(req *tpb.Wallets, stream tpb.TransactionService_GetUserTransactionsServer) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	transactions, err := s.transactionsRepository.GetTransactionsByWalletIDs(timeout, nil, req.GetWalletId())
	if err != nil {
		return err
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
			return err
		}
	}

	return nil
}
