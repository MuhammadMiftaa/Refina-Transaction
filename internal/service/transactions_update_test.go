package service

import (
	"context"
	"errors"
	"testing"

	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =====================================================================
// UpdateTransaction
// =====================================================================

func TestUpdateTransaction_SuccessAmountChange(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel() // amount=50000, expense
	updated := existing
	updated.Amount = 75000

	wallet := sampleWalletProto(walletTestID, 150000)
	updatedWallet := sampleWalletProto(walletTestID, 125000)

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      75000,
		Date:        txnFixTime,
		Description: "Makan siang updated",
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	// amount changed — fetch wallet and update balance
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, txnTestID.String(), result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_SuccessSameAmountAndWallet(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel()
	updated := existing
	updated.Description = "Updated description"

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000, // same amount
		Date:        txnFixTime,
		Description: "Updated description",
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	// no wallet/balance update since amount is the same
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, txnTestID.String(), result.ID)
	d.walletClient.AssertNotCalled(t, "GetWalletByID")
	d.assertAll(t)
}

func TestUpdateTransaction_SuccessWalletChange(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	newWalletID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	existing := sampleTransactionModel() // walletTestID, expense, 50000
	updated := existing
	updated.WalletID = newWalletID

	oldWallet := sampleWalletProto(walletTestID, 150000)
	newWallet := sampleWalletProto(newWalletID, 300000)
	updatedOldWallet := sampleWalletProto(walletTestID, 200000)
	updatedNewWallet := sampleWalletProto(newWalletID, 250000)

	req := dto.TransactionsRequest{
		WalletID:    newWalletID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: "Moved to new wallet",
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	// wallet changed — restore old wallet, deduct new wallet
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(oldWallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.MatchedBy(func(w interface{ GetId() string }) bool {
		return w.GetId() == walletTestID.String()
	})).Return(updatedOldWallet, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, newWalletID.String()).Return(newWallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.MatchedBy(func(w interface{ GetId() string }) bool {
		return w.GetId() == newWalletID.String()
	})).Return(updatedNewWallet, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_SuccessCategoryChange(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	newCatID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	existing := sampleTransactionModel()
	newCat := model.Categories{Base: model.Base{ID: newCatID}, Name: "Transportasi", Type: model.Expense}
	updated := existing
	updated.CategoryID = newCatID

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  newCatID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: existing.Description,
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.categoryRepo.On("GetCategoryByID", mock.Anything, d.tx, newCatID.String()).Return(newCat, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	d.walletClient.AssertNotCalled(t, "GetWalletByID")
	d.assertAll(t)
}

func TestUpdateTransaction_NotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, "bad-id").
		Return(model.Transactions{}, errors.New("transaction not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), "bad-id", dto.TransactionsRequest{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_InvalidNewCategoryID(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel()
	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  "not-a-uuid",
		Amount:      50000,
		Date:        txnFixTime,
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.categoryRepo.On("GetCategoryByID", mock.Anything, d.tx, "not-a-uuid").
		Return(model.Categories{}, errors.New("record not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_WithAttachmentCreate(t *testing.T) {
	t.Skip("UploadAttachment requires a real MinIO instance; covered by integration tests")
}

func TestUpdateTransaction_WithAttachmentDelete(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	attachmentIDToDelete := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	existing := sampleTransactionModel()
	updated := existing

	attToDelete := model.Attachments{
		Base:          model.Base{ID: attachmentIDToDelete},
		TransactionID: txnTestID,
		Image:         "http://minio/old.jpg",
	}

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: existing.Description,
		Attachments: []dto.UpdateAttachmentsRequest{
			{Status: "delete", Files: []string{attachmentIDToDelete.String()}},
		},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, d.tx, attachmentIDToDelete.String()).Return(attToDelete, nil)
	d.attachmentRepo.On("DeleteAttachment", mock.Anything, d.tx, attToDelete).Return(attToDelete, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_WithAttachmentDeleteNotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel()
	updated := existing

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: existing.Description,
		Attachments: []dto.UpdateAttachmentsRequest{
			{Status: "delete", Files: []string{"nonexistent-att-id"}},
		},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, d.tx, "nonexistent-att-id").
		Return(model.Attachments{}, errors.New("record not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_WithAttachmentInvalidStatus(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel()
	updated := existing

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: existing.Description,
		Attachments: []dto.UpdateAttachmentsRequest{
			{Status: "invalid_status", Files: []string{"some-id"}},
		},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid attachment status")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_OutboxError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel()
	updated := existing

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: "No change",
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(errors.New("outbox error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateTransaction_CommitError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	existing := sampleTransactionModel()
	updated := existing

	req := dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: "No change",
		Attachments: []dto.UpdateAttachmentsRequest{},
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(existing, nil)
	d.transactionRepo.On("UpdateTransaction", mock.Anything, d.tx, mock.Anything).Return(updated, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(errors.New("commit error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.UpdateTransaction(context.Background(), txnTestID.String(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}
