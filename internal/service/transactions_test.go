package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"refina-transaction/internal/repository"
	"refina-transaction/internal/service/mocks"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"

	wpb "github.com/MuhammadMiftaa/Refina-Protobuf/wallet"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// Test Dependency Container
// ─────────────────────────────────────────────

type transactionTestDeps struct {
	txManager      *mocks.MockTxManager
	transactionRepo *mocks.MockTransactionsRepository
	categoryRepo   *mocks.MockCategoriesRepository
	attachmentRepo *mocks.MockAttachmentsRepository
	outboxRepo     *mocks.MockOutboxRepository
	walletClient   *mocks.MockWalletClient
	tx             *mocks.MockTransaction
}

func newTransactionTestDeps() *transactionTestDeps {
	return &transactionTestDeps{
		txManager:      new(mocks.MockTxManager),
		transactionRepo: new(mocks.MockTransactionsRepository),
		categoryRepo:   new(mocks.MockCategoriesRepository),
		attachmentRepo: new(mocks.MockAttachmentsRepository),
		outboxRepo:     new(mocks.MockOutboxRepository),
		walletClient:   new(mocks.MockWalletClient),
		tx:             new(mocks.MockTransaction),
	}
}

func (d *transactionTestDeps) service() TransactionsService {
	// We pass nil for minio since file upload tests would need a real/mocked minio.
	// Tests that require upload should skip or use integration tests.
	return NewTransactionService(
		d.txManager,
		d.transactionRepo,
		d.walletClient,
		d.categoryRepo,
		d.attachmentRepo,
		d.outboxRepo,
		nil, // minio — nil is acceptable for non-upload tests
	)
}

func (d *transactionTestDeps) assertAll(t *testing.T) {
	t.Helper()
	d.txManager.AssertExpectations(t)
	d.transactionRepo.AssertExpectations(t)
	d.categoryRepo.AssertExpectations(t)
	d.attachmentRepo.AssertExpectations(t)
	d.outboxRepo.AssertExpectations(t)
	d.walletClient.AssertExpectations(t)
	d.tx.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Fixed UUIDs & Timestamps
// ─────────────────────────────────────────────

var (
	txnTestID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	walletTestID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	catTestID    = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	wallet2ID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	cashInCatID  = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	cashOutCatID = uuid.MustParse("00000000-0000-0000-0000-000000000012")
	txnFixTime   = time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
)

// ─────────────────────────────────────────────
// Sample Data Factories
// ─────────────────────────────────────────────

func sampleExpenseCategory() model.Categories {
	return model.Categories{
		Base: model.Base{ID: catTestID},
		Name: "Makanan",
		Type: model.Expense,
	}
}

func sampleIncomeCategory() model.Categories {
	return model.Categories{
		Base: model.Base{ID: catTestID},
		Name: "Gaji Bulanan",
		Type: model.Income,
	}
}

func sampleFundTransferCashOut() model.Categories {
	return model.Categories{
		Base: model.Base{ID: cashOutCatID},
		Name: "Cash Out",
		Type: model.FundTransfer,
	}
}

func sampleFundTransferCashIn() model.Categories {
	return model.Categories{
		Base: model.Base{ID: cashInCatID},
		Name: "Cash In",
		Type: model.FundTransfer,
	}
}

func sampleTransactionModel() model.Transactions {
	return model.Transactions{
		Base:            model.Base{ID: txnTestID, CreatedAt: txnFixTime, UpdatedAt: txnFixTime},
		WalletID:        walletTestID,
		CategoryID:      catTestID,
		Amount:          50000,
		TransactionDate: txnFixTime,
		Description:     "Makan siang",
		Category:        sampleExpenseCategory(),
		Attachments:     []model.Attachments{},
	}
}

func sampleWalletProto(id uuid.UUID, balance float64) *wpb.Wallet {
	return &wpb.Wallet{
		Id:      id.String(),
		Name:    "BCA Tabungan",
		Balance: balance,
	}
}

func sampleTransactionRequest() dto.TransactionsRequest {
	return dto.TransactionsRequest{
		WalletID:    walletTestID.String(),
		CategoryID:  catTestID.String(),
		Amount:      50000,
		Date:        txnFixTime,
		Description: "Makan siang",
		Attachments: []dto.UpdateAttachmentsRequest{},
	}
}

// =====================================================================
// GetAllTransactions
// =====================================================================

func TestGetAllTransactions_Success(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txns := []model.Transactions{sampleTransactionModel()}
	d.transactionRepo.On("GetAllTransactions", mock.Anything, nil).Return(txns, nil)

	result, err := svc.GetAllTransactions(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, txnTestID.String(), result[0].ID)
	assert.Equal(t, walletTestID.String(), result[0].WalletID)
	assert.Equal(t, float64(50000), result[0].Amount)
	d.assertAll(t)
}

func TestGetAllTransactions_EmptyList(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	d.transactionRepo.On("GetAllTransactions", mock.Anything, nil).Return([]model.Transactions{}, nil)

	result, err := svc.GetAllTransactions(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
	d.assertAll(t)
}

func TestGetAllTransactions_RepositoryError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	d.transactionRepo.On("GetAllTransactions", mock.Anything, nil).
		Return([]model.Transactions{}, errors.New("db error"))

	result, err := svc.GetAllTransactions(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get all transactions")
	d.assertAll(t)
}

// =====================================================================
// GetTransactionByID
// =====================================================================

func TestGetTransactionByID_Success(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	d.transactionRepo.On("GetTransactionByID", mock.Anything, nil, txnTestID.String()).Return(txn, nil)
	d.attachmentRepo.On("GetAttachmentsByTransactionID", mock.Anything, nil, txnTestID.String()).
		Return([]model.Attachments{}, nil)

	result, err := svc.GetTransactionByID(context.Background(), txnTestID.String())

	assert.NoError(t, err)
	assert.Equal(t, txnTestID.String(), result.ID)
	assert.Equal(t, "Makan siang", result.Description)
	assert.Equal(t, "Makanan", result.CategoryName)
	assert.Empty(t, result.Attachments)
	d.assertAll(t)
}

func TestGetTransactionByID_WithAttachments(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	attModel := model.Attachments{
		Base:          model.Base{ID: attID},
		TransactionID: txnTestID,
		Image:         "http://minio/bucket/receipt.jpg",
		Format:        ".jpg",
		Size:          2048,
	}
	d.transactionRepo.On("GetTransactionByID", mock.Anything, nil, txnTestID.String()).Return(txn, nil)
	d.attachmentRepo.On("GetAttachmentsByTransactionID", mock.Anything, nil, txnTestID.String()).
		Return([]model.Attachments{attModel}, nil)

	result, err := svc.GetTransactionByID(context.Background(), txnTestID.String())

	assert.NoError(t, err)
	assert.Len(t, result.Attachments, 1)
	assert.Equal(t, "http://minio/bucket/receipt.jpg", result.Attachments[0].Image)
	d.assertAll(t)
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	d.transactionRepo.On("GetTransactionByID", mock.Anything, nil, "bad-id").
		Return(model.Transactions{}, errors.New("transaction not found"))

	result, err := svc.GetTransactionByID(context.Background(), "bad-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestGetTransactionByID_AttachmentRepositoryError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	d.transactionRepo.On("GetTransactionByID", mock.Anything, nil, txnTestID.String()).Return(txn, nil)
	d.attachmentRepo.On("GetAttachmentsByTransactionID", mock.Anything, nil, txnTestID.String()).
		Return([]model.Attachments{}, errors.New("db error"))

	result, err := svc.GetTransactionByID(context.Background(), txnTestID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get attachments")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

// =====================================================================
// GetTransactionsByWalletIDs
// =====================================================================

func TestGetTransactionsByWalletIDs_Success(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	ids := []string{walletTestID.String()}
	txns := []model.Transactions{sampleTransactionModel()}
	d.transactionRepo.On("GetTransactionsByWalletIDs", mock.Anything, nil, ids).Return(txns, nil)

	result, err := svc.GetTransactionsByWalletIDs(context.Background(), ids)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	d.assertAll(t)
}

func TestGetTransactionsByWalletIDs_EmptyResult(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	ids := []string{walletTestID.String()}
	d.transactionRepo.On("GetTransactionsByWalletIDs", mock.Anything, nil, ids).
		Return([]model.Transactions{}, nil)

	result, err := svc.GetTransactionsByWalletIDs(context.Background(), ids)

	assert.NoError(t, err)
	assert.Empty(t, result)
	d.assertAll(t)
}

func TestGetTransactionsByWalletIDs_RepositoryError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	ids := []string{walletTestID.String()}
	d.transactionRepo.On("GetTransactionsByWalletIDs", mock.Anything, nil, ids).
		Return([]model.Transactions{}, errors.New("db error"))

	result, err := svc.GetTransactionsByWalletIDs(context.Background(), ids)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get transactions by wallet ids")
	d.assertAll(t)
}

// =====================================================================
// GetTransactionsByCursor
// =====================================================================

func TestGetTransactionsByCursor_Success(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	q := repository.CursorQuery{
		WalletIDs: []string{walletTestID.String()},
		PageSize:  10,
	}
	txns := []model.Transactions{sampleTransactionModel()}
	d.transactionRepo.On("GetTransactionsByCursor", mock.Anything, nil, q).
		Return(txns, int64(1), nil)

	result, total, err := svc.GetTransactionsByCursor(context.Background(), q)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	d.assertAll(t)
}

func TestGetTransactionsByCursor_RepositoryError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	q := repository.CursorQuery{WalletIDs: []string{walletTestID.String()}}
	d.transactionRepo.On("GetTransactionsByCursor", mock.Anything, nil, q).
		Return([]model.Transactions{}, int64(0), errors.New("db error"))

	result, total, err := svc.GetTransactionsByCursor(context.Background(), q)

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get transactions by cursor")
	d.assertAll(t)
}
