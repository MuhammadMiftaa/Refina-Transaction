package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"refina-transaction/config/log"
	"refina-transaction/config/miniofs"
	"refina-transaction/interface/grpc/client"
	"refina-transaction/internal/repository"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"
	helper "refina-transaction/internal/utils"
	"refina-transaction/internal/utils/data"

	"github.com/google/uuid"
)

type TransactionsService interface {
	GetAllTransactions(ctx context.Context) ([]dto.TransactionsResponse, error)
	GetTransactionByID(ctx context.Context, id string) (dto.TransactionsResponse, error)
	GetTransactionsByWalletIDs(ctx context.Context, ids []string) ([]dto.TransactionsResponse, error)
	CreateTransaction(ctx context.Context, transaction dto.TransactionsRequest) (dto.TransactionsResponse, error)
	FundTransfer(ctx context.Context, transaction dto.FundTransferRequest) (dto.FundTransferResponse, error)
	UploadAttachment(ctx context.Context, transactionID string, files []string) ([]dto.AttachmentsResponse, error)
	UpdateTransaction(ctx context.Context, id string, transaction dto.TransactionsRequest) (dto.TransactionsResponse, error)
	DeleteTransaction(ctx context.Context, id string) (dto.TransactionsResponse, error)
}

type transactionsService struct {
	txManager        repository.TxManager
	transactionRepo  repository.TransactionsRepository
	categoryRepo     repository.CategoriesRepository
	attachmentRepo   repository.AttachmentsRepository
	outboxRepository repository.OutboxRepository
	minio            *miniofs.MinIOManager
	walletClient     client.WalletClient
}

func NewTransactionService(txManager repository.TxManager, transactionRepo repository.TransactionsRepository, walletRepo client.WalletClient, categoryRepo repository.CategoriesRepository, attachmentRepo repository.AttachmentsRepository, outboxRepository repository.OutboxRepository, minio *miniofs.MinIOManager) TransactionsService {
	return &transactionsService{
		txManager:        txManager,
		transactionRepo:  transactionRepo,
		categoryRepo:     categoryRepo,
		attachmentRepo:   attachmentRepo,
		outboxRepository: outboxRepository,
		minio:            minio,
		walletClient:     walletRepo,
	}
}

func (transaction_serv *transactionsService) GetAllTransactions(ctx context.Context) ([]dto.TransactionsResponse, error) {
	transactions, err := transaction_serv.transactionRepo.GetAllTransactions(ctx, nil)
	if err != nil {
		return nil, errors.New("failed to get transactions")
	}

	transactionResponses := make([]dto.TransactionsResponse, 0, len(transactions))
	for _, transaction := range transactions {
		transactionResponse := helper.ConvertToResponseType(transaction).(dto.TransactionsResponse)
		transactionResponses = append(transactionResponses, transactionResponse)
	}

	return transactionResponses, nil
}

func (transaction_serv *transactionsService) GetTransactionByID(ctx context.Context, id string) (dto.TransactionsResponse, error) {
	transaction, err := transaction_serv.transactionRepo.GetTransactionByID(ctx, nil, id)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("transaction not found")
	}

	transactionResponse := helper.ConvertToResponseType(transaction).(dto.TransactionsResponse)

	attachments, err := transaction_serv.attachmentRepo.GetAttachmentsByTransactionID(ctx, nil, transaction.ID.String())
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to get attachments")
	}

	if len(attachments) > 0 {
		for _, attachment := range attachments {
			if attachment.Image != "" {
				result := dto.AttachmentsResponse{
					ID:            attachment.ID.String(),
					TransactionID: attachment.TransactionID.String(),
					Image:         attachment.Image,
					Format:        attachment.Format,
					Size:          attachment.Size,
				}
				transactionResponse.Attachments = append(transactionResponse.Attachments, result)
			}
		}
	} else {
		transactionResponse.Attachments = make([]dto.AttachmentsResponse, 0, len(attachments))
	}

	return transactionResponse, nil
}

func (transaction_serv *transactionsService) GetTransactionsByWalletIDs(ctx context.Context, ids []string) ([]dto.TransactionsResponse, error) {
	transactions, err := transaction_serv.transactionRepo.GetTransactionsByWalletIDs(ctx, nil, ids)
	if err != nil {
		return nil, errors.New("failed to get transactions")
	}

	transactionResponses := make([]dto.TransactionsResponse, 0, len(transactions))
	for _, transaction := range transactions {
		transactionResponse := helper.ConvertToResponseType(transaction).(dto.TransactionsResponse)
		transactionResponses = append(transactionResponses, transactionResponse)
	}

	return transactionResponses, nil
}

func (transaction_serv *transactionsService) CreateTransaction(ctx context.Context, transaction dto.TransactionsRequest) (dto.TransactionsResponse, error) {
	tx, err := transaction_serv.txManager.Begin(ctx)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to create transaction")
	}

	defer func() {
		// Rollback otomatis jika transaksi belum di-commit
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	// Check if wallet and category exist
	wallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transaction.WalletID)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("wallet not found")
	}

	category, err := transaction_serv.categoryRepo.GetCategoryByID(ctx, tx, transaction.CategoryID)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("category not found")
	}

	// Check if wallet has sufficient balance
	if wallet.GetBalance() < transaction.Amount {
		return dto.TransactionsResponse{}, errors.New("insufficient wallet balance")
	}

	// Check if transaction type is valid and update wallet balance
	switch category.Type {
	case "expense":
		wallet.Balance -= transaction.Amount
	case "income":
		wallet.Balance += transaction.Amount
	default:
		return dto.TransactionsResponse{}, errors.New("invalid transaction type")
	}

	// Parse ID from JSON to valid UUID
	CategoryID, err := helper.ParseUUID(transaction.CategoryID)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("invalid category id")
	}

	WalletID, err := helper.ParseUUID(transaction.WalletID)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("invalid wallet id")
	}

	// Update wallet balance
	_, err = transaction_serv.walletClient.UpdateWallet(ctx, wallet)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to update wallet")
	}

	// Create transaction
	transactionNew, err := transaction_serv.transactionRepo.CreateTransaction(ctx, tx, model.Transactions{
		WalletID:        WalletID,
		CategoryID:      CategoryID,
		Amount:          transaction.Amount,
		TransactionDate: transaction.Date,
		Description:     transaction.Description,
	})
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to create transaction")
	}

	// ? If attachments exist, upload attachments
	if len(transaction.Attachments) > 0 {
		for _, attachment := range transaction.Attachments {
			// * Create new attachment
			if len(attachment.Files) == 0 {
				return dto.TransactionsResponse{}, errors.New("no files to upload")
			}

			if _, err := transaction_serv.UploadAttachment(ctx, transactionNew.ID.String(), attachment.Files); err != nil {
				return dto.TransactionsResponse{}, fmt.Errorf("failed to upload attachment: %w", err)
			}
		}
	}

	transactionResponse := helper.ConvertToResponseType(transactionNew).(dto.TransactionsResponse)

	payload, err := json.Marshal(transactionResponse)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to marshal transaction response")
	}

	outboxMsg := &model.OutboxMessage{
		AggregateID: transactionResponse.ID,
		EventType:   data.OUTBOX_EVENT_TRANSACTION_CREATED,
		Payload:     payload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}

	if err := transaction_serv.outboxRepository.Create(ctx, tx, outboxMsg); err != nil {
		tx.Rollback()
		return dto.TransactionsResponse{}, err
	}

	// Commit transaksi jika semua sukses
	if err := tx.Commit(); err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to commit transaction")
	}

	return transactionResponse, nil
}

func (transaction_serv *transactionsService) FundTransfer(ctx context.Context, transaction dto.FundTransferRequest) (dto.FundTransferResponse, error) {
	tx, err := transaction_serv.txManager.Begin(ctx)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to create transaction")
	}

	defer func() {
		// Rollback otomatis jika transaksi belum di-commit
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	// Check if wallet and category exist
	fromWallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transaction.FromWalletID)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("source wallet not found")
	}

	toWallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transaction.ToWalletID)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("destination wallet not found")
	}

	// Check if wallet has sufficient balance
	if fromWallet.GetBalance() < (transaction.Amount + transaction.AdminFee) {
		return dto.FundTransferResponse{}, errors.New("insufficient wallet balance")
	}

	// Check if source and destination wallets are the same
	if fromWallet.GetId() == toWallet.GetId() {
		return dto.FundTransferResponse{}, errors.New("source wallet and destination wallet cannot be the same")
	}

	fromWallet.Balance -= (transaction.Amount + transaction.AdminFee)
	toWallet.Balance += transaction.Amount

	// Parse ID from JSON to valid UUID
	FromWalletID, err := helper.ParseUUID(transaction.FromWalletID)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("invalid from wallet id")
	}

	ToWalletID, err := helper.ParseUUID(transaction.ToWalletID)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("invalid to wallet id")
	}

	// Parse CategoryID from JSON to valid UUID
	FromCategoryID, err := helper.ParseUUID(transaction.CashOutCategoryID)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("invalid from category id")
	}

	ToCategoryID, err := helper.ParseUUID(transaction.CashInCategoryID)
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("invalid to category id")
	}

	// Update wallet balance
	if _, err = transaction_serv.walletClient.UpdateWallet(ctx, fromWallet); err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to update from wallet")
	}
	if _, err = transaction_serv.walletClient.UpdateWallet(ctx, toWallet); err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to update to wallet")
	}

	transactionNewFrom, err := transaction_serv.transactionRepo.CreateTransaction(ctx, tx, model.Transactions{
		WalletID:        FromWalletID,
		CategoryID:      FromCategoryID,
		Amount:          transaction.Amount + transaction.AdminFee,
		TransactionDate: transaction.Date,
		Description:     "fund transfer to " + toWallet.GetName() + "(Cash Out)",
	})
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to create from transaction")
	}

	transactionNewTo, err := transaction_serv.transactionRepo.CreateTransaction(ctx, tx, model.Transactions{
		WalletID:        ToWalletID,
		CategoryID:      ToCategoryID,
		Amount:          transaction.Amount,
		TransactionDate: transaction.Date,
		Description:     "fund transfer from " + fromWallet.GetName() + "(Cash In)",
	})
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to create to transaction")
	}

	transactionNewFromPayload, err := json.Marshal(helper.ConvertToResponseType(transactionNewFrom).(dto.TransactionsResponse))
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to marshal transaction response")
	}

	transactionNewToPayload, err := json.Marshal(helper.ConvertToResponseType(transactionNewTo).(dto.TransactionsResponse))
	if err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to marshal transaction response")
	}

	if err := transaction_serv.outboxRepository.Create(ctx, tx, &model.OutboxMessage{
		AggregateID: transactionNewFrom.ID.String(),
		EventType:   data.OUTBOX_EVENT_TRANSACTION_CREATED,
		Payload:     transactionNewFromPayload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}); err != nil {
		tx.Rollback()
		return dto.FundTransferResponse{}, err
	}

	if err := transaction_serv.outboxRepository.Create(ctx, tx, &model.OutboxMessage{
		AggregateID: transactionNewTo.ID.String(),
		EventType:   data.OUTBOX_EVENT_TRANSACTION_CREATED,
		Payload:     transactionNewToPayload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}); err != nil {
		tx.Rollback()
		return dto.FundTransferResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return dto.FundTransferResponse{}, errors.New("failed to commit transaction")
	}

	response := dto.FundTransferResponse{
		CashOutTransactionID: transactionNewFrom.ID.String(),
		CashInTransactionID:  transactionNewTo.ID.String(),
		FromWalletID:         transaction.FromWalletID,
		ToWalletID:           transaction.ToWalletID,
		Amount:               transaction.Amount,
		Date:                 transaction.Date,
		Description:          transaction.Description,
	}

	return response, nil
}

func (transaction_serv *transactionsService) UploadAttachment(ctx context.Context, transactionID string, files []string) ([]dto.AttachmentsResponse, error) {
	var attachmentResponses []dto.AttachmentsResponse

	if transactionID == "" {
		log.Error("transaction ID is required")
		return nil, errors.New("transaction ID is required")
	}
	if len(files) == 0 {
		log.Error("no files to upload")
		return nil, errors.New("no files to upload")
	}

	for idx, file := range files {
		if file == "" {
			log.Error("file is empty")
			return nil, errors.New("file is empty")
		}

		ctx := context.Background()
		fileReq := miniofs.UploadRequest{
			Prefix:     fmt.Sprintf("%s_%s", miniofs.TRANSACTION_ATTACHMENT_PREFIX, transactionID),
			Base64Data: file,
			BucketName: miniofs.TRANSACTION_ATTACHMENT_BUCKET,
			Validation: miniofs.CreateImageValidationConfig(),
		}
		res, err := transaction_serv.minio.UploadFile(ctx, fileReq)
		if err != nil {
			log.Error(fmt.Sprintf("failed to upload file %d: %v", idx+1, err))
			return nil, errors.New("failed to upload file")
		}

		// Save attachment to database
		TransactionUUID, err := uuid.Parse(transactionID)
		if err != nil {
			log.Error(fmt.Sprintf("invalid transaction ID %s: %v", transactionID, err))
			return nil, errors.New("invalid transaction id")
		}

		attachment, err := transaction_serv.attachmentRepo.CreateAttachment(ctx, nil, model.Attachments{
			Image:         res.URL,
			TransactionID: TransactionUUID,
			Size:          res.Size,
			Format:        res.Ext,
		})
		if err != nil {
			log.Error(fmt.Sprintf("failed to create attachment for transaction %s: %v", transactionID, err))
			return nil, errors.New("failed to create attachment")
		}

		attachmentResponse := dto.AttachmentsResponse{
			ID:            attachment.ID.String(),
			Image:         attachment.Image,
			TransactionID: attachment.TransactionID.String(),
			CreatedAt:     attachment.CreatedAt.String(),
		}

		attachmentResponses = append(attachmentResponses, attachmentResponse)
	}

	return attachmentResponses, nil
}

func (transaction_serv *transactionsService) UpdateTransaction(ctx context.Context, id string, transaction dto.TransactionsRequest) (dto.TransactionsResponse, error) {
	// ! Begin a new transaction
	tx, err := transaction_serv.txManager.Begin(ctx)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to create transaction")
	}

	// ! Defer rollback if there is an error
	defer func() {
		// Rollback otomatis jika transaksi belum di-commit
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	// ? Check if transaction exist
	transactionExist, err := transaction_serv.transactionRepo.GetTransactionByID(ctx, tx, id)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("transaction not found")
	}

	// ? If category ID is different, update category
	if transaction.CategoryID != transactionExist.CategoryID.String() {
		CategoryID, err := helper.ParseUUID(transaction.CategoryID)
		if err != nil {
			return dto.TransactionsResponse{}, errors.New("invalid category id")
		}

		// * Check if category exist
		_, err = transaction_serv.categoryRepo.GetCategoryByID(ctx, tx, transaction.CategoryID)
		if err != nil {
			return dto.TransactionsResponse{}, errors.New("category not found")
		}

		transactionExist.CategoryID = CategoryID
	}

	// ? If wallet ID is different, update wallet balance
	if transaction.WalletID != transactionExist.WalletID.String() {
		// *  Check if wallet exist
		oldWallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transactionExist.WalletID.String())
		if err != nil {
			return dto.TransactionsResponse{}, errors.New("wallet not found")
		}

		// *  Update wallet balance
		switch transactionExist.Category.Type {
		case "expense":
			oldWallet.Balance += transactionExist.Amount
		case "income":
			oldWallet.Balance -= transactionExist.Amount
		default:
			return dto.TransactionsResponse{}, errors.New("invalid transaction type")
		}

		if _, err = transaction_serv.walletClient.UpdateWallet(ctx, oldWallet); err != nil {
			return dto.TransactionsResponse{}, errors.New("failed to update wallet")
		}

		// *  Check if new wallet exist
		newWallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transaction.WalletID)
		if err != nil {
			return dto.TransactionsResponse{}, errors.New("new wallet not found")
		}

		// *  Update wallet balance
		switch transactionExist.Category.Type {
		case "expense":
			newWallet.Balance -= transaction.Amount
		case "income":
			newWallet.Balance += transaction.Amount
		default:
			return dto.TransactionsResponse{}, errors.New("invalid transaction type")
		}

		if _, err = transaction_serv.walletClient.UpdateWallet(ctx, newWallet); err != nil {
			return dto.TransactionsResponse{}, errors.New("failed to update new wallet")
		}

		// *  Parse ID from JSON to valid UUID
		WalletID, err := helper.ParseUUID(transaction.WalletID)
		if err != nil {
			return dto.TransactionsResponse{}, errors.New("invalid wallet id")
		}
		transactionExist.WalletID = WalletID
	}

	// ? Update transaction fields
	if transaction.Amount != transactionExist.Amount {
		// *  Update wallet balance
		oldWallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transactionExist.WalletID.String())
		if err != nil {
			return dto.TransactionsResponse{}, errors.New("wallet not found")
		}

		// *  Update wallet balance
		switch transactionExist.Category.Type {
		case "expense":
			oldWallet.Balance += transactionExist.Amount
			oldWallet.Balance -= transaction.Amount
		case "income":
			oldWallet.Balance -= transactionExist.Amount
			oldWallet.Balance += transaction.Amount
		default:
			return dto.TransactionsResponse{}, errors.New("invalid transaction type")
		}

		if _, err = transaction_serv.walletClient.UpdateWallet(ctx, oldWallet); err != nil {
			return dto.TransactionsResponse{}, errors.New("failed to update wallet")
		}

		// *  Update transaction amount
		transactionExist.Amount = transaction.Amount
	}

	// ? Update transaction date
	if !transaction.Date.IsZero() {
		transactionExist.TransactionDate = transaction.Date
	}

	// ? Update description
	if transaction.Description != "" {
		transactionExist.Description = transaction.Description
	}

	// ? Update transaction
	transactionUpdated, err := transaction_serv.transactionRepo.UpdateTransaction(ctx, tx, transactionExist)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to update transaction")
	}

	// ? If attachments exist, update attachments
	if len(transaction.Attachments) > 0 {
		for _, attachment := range transaction.Attachments {
			switch attachment.Status {
			case "create":
				// * Create new attachment
				if len(attachment.Files) == 0 {
					return dto.TransactionsResponse{}, errors.New("no files to upload")
				}

				if _, err := transaction_serv.UploadAttachment(ctx, transactionUpdated.ID.String(), attachment.Files); err != nil {
					return dto.TransactionsResponse{}, fmt.Errorf("failed to upload attachment: %w", err)
				}

			case "delete":
				// * Delete attachment
				if len(attachment.Files) == 0 {
					return dto.TransactionsResponse{}, errors.New("no files to delete")
				}

				for _, ID := range attachment.Files {
					// * Get attachment by ID
					attachmentToDelete, err := transaction_serv.attachmentRepo.GetAttachmentByID(ctx, tx, ID)
					if err != nil {
						return dto.TransactionsResponse{}, fmt.Errorf("attachment with file %s not found: %w", ID, err)
					}

					// * Check if attachment belongs to transaction
					if attachmentToDelete.TransactionID != transactionUpdated.ID {
						return dto.TransactionsResponse{}, fmt.Errorf("attachment with file %s does not belong to transaction %s", ID, transactionUpdated.ID)
					}

					// * Delete file from database
					if _, err := transaction_serv.attachmentRepo.DeleteAttachment(ctx, tx, attachmentToDelete); err != nil {
						return dto.TransactionsResponse{}, fmt.Errorf("attachment with file %v not found: %w", attachmentToDelete, err)
					}
				}

			default:
				return dto.TransactionsResponse{}, errors.New("invalid attachment status")
			}
		}
	}

	transactionResponse := helper.ConvertToResponseType(transactionUpdated).(dto.TransactionsResponse)

	payload, err := json.Marshal(transactionResponse)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to marshal transaction response")
	}

	outboxMsg := &model.OutboxMessage{
		AggregateID: transactionResponse.ID,
		EventType:   data.OUTBOX_EVENT_TRANSACTION_UPDATED,
		Payload:     payload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}

	if err := transaction_serv.outboxRepository.Create(ctx, tx, outboxMsg); err != nil {
		tx.Rollback()
		return dto.TransactionsResponse{}, err
	}

	// ! Commit transaction if all operations are successful
	if err = tx.Commit(); err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to commit transaction")
	}

	return transactionResponse, nil
}

func (transaction_serv *transactionsService) DeleteTransaction(ctx context.Context, id string) (dto.TransactionsResponse, error) {
	tx, err := transaction_serv.txManager.Begin(ctx)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to create transaction")
	}

	defer func() {
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	// Check if transaction exist
	transactionExist, err := transaction_serv.transactionRepo.GetTransactionByID(ctx, tx, id)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("transaction not found")
	}

	// Get wallet to update balance
	wallet, err := transaction_serv.walletClient.GetWalletByID(ctx, transactionExist.WalletID.String())
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("wallet not found")
	}

	// Update wallet balance
	if transactionExist.Category.Type == "expense" {
		wallet.Balance += transactionExist.Amount
	} else if transactionExist.Category.Type == "income" {
		wallet.Balance -= transactionExist.Amount
	} else {
		if transactionExist.Category.Name == "Cash Out" {
			wallet.Balance += transactionExist.Amount
		} else if transactionExist.Category.Name == "Cash In" {
			wallet.Balance -= transactionExist.Amount
		} else {
			return dto.TransactionsResponse{}, errors.New("invalid transaction type")
		}
	}

	// Update wallet balance
	_, err = transaction_serv.walletClient.UpdateWallet(ctx, wallet)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to update wallet")
	}

	// Delete transaction
	transactionDeleted, err := transaction_serv.transactionRepo.DeleteTransaction(ctx, tx, transactionExist)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to delete transaction")
	}

	transactionResponse := helper.ConvertToResponseType(transactionDeleted).(dto.TransactionsResponse)

	payload, err := json.Marshal(transactionResponse)
	if err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to marshal transaction response")
	}

	outboxMsg := &model.OutboxMessage{
		AggregateID: transactionResponse.ID,
		EventType:   data.OUTBOX_EVENT_TRANSACTION_DELETED,
		Payload:     payload,
		Published:   false,
		MaxRetries:  data.OUTBOX_PUBLISH_MAX_RETRIES,
	}

	if err := transaction_serv.outboxRepository.Create(ctx, tx, outboxMsg); err != nil {
		tx.Rollback()
		return dto.TransactionsResponse{}, err
	}

	// Commit transaksi jika semua sukses
	if err := tx.Commit(); err != nil {
		return dto.TransactionsResponse{}, errors.New("failed to commit transaction")
	}

	return transactionResponse, nil
}
