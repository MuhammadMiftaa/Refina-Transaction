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
// CreateTransaction
// =====================================================================

func TestCreateTransaction_SuccessExpense(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest() // expense, amount=50000
	cat := sampleExpenseCategory()
	wallet := sampleWalletProto(walletTestID, 200000)
	updatedWallet := sampleWalletProto(walletTestID, 150000)
	createdTxn := sampleTransactionModel()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.Anything).Return(createdTxn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, txnTestID.String(), result.ID)
	assert.Equal(t, walletTestID.String(), result.WalletID)
	assert.Equal(t, float64(50000), result.Amount)
	d.assertAll(t)
}

func TestCreateTransaction_SuccessIncome(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	req.Amount = 5000000

	cat := sampleIncomeCategory()
	wallet := sampleWalletProto(walletTestID, 100000)
	updatedWallet := sampleWalletProto(walletTestID, 5100000)
	createdTxn := sampleTransactionModel()
	createdTxn.Amount = 5000000
	createdTxn.Category = cat

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.Anything).Return(createdTxn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_IsWalletNotCreated_SkipsWalletCheck(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	req.IsWalletNotCreated = true

	cat := sampleExpenseCategory()
	createdTxn := sampleTransactionModel()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.Anything).Return(createdTxn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, txnTestID.String(), result.ID)
	// wallet client should NOT be called
	d.walletClient.AssertNotCalled(t, "GetWalletByID")
	d.walletClient.AssertNotCalled(t, "UpdateWallet")
	d.assertAll(t)
}

func TestCreateTransaction_InvalidCategoryID(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	req.CategoryID = "not-a-uuid"

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, "not-a-uuid").
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_CategoryNotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).
		Return(model.Categories{}, errors.New("record not found"))

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_InvalidWalletID(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	req.WalletID = "not-a-uuid"

	cat := sampleExpenseCategory()
	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid wallet id")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_WalletNotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := sampleExpenseCategory()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).
		Return(nil, errors.New("wallet not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wallet not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_InsufficientBalance(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest() // amount = 50000
	cat := sampleExpenseCategory()
	wallet := sampleWalletProto(walletTestID, 1000) // balance less than amount

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient wallet balance")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_InvalidCategoryType(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := model.Categories{
		Base: model.Base{ID: catTestID},
		Name: "Unknown",
		Type: model.FundTransfer, // fund_transfer without explicit handling
	}
	wallet := sampleWalletProto(walletTestID, 200000)

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction type")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_UpdateWalletError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := sampleExpenseCategory()
	wallet := sampleWalletProto(walletTestID, 200000)

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(nil, errors.New("grpc error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update wallet balance")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_BeginTxError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := sampleExpenseCategory()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(nil, errors.New("begin tx error"))

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_InsertDBError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := sampleExpenseCategory()
	wallet := sampleWalletProto(walletTestID, 200000)
	updatedWallet := sampleWalletProto(walletTestID, 150000)

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.Anything).
		Return(model.Transactions{}, errors.New("db insert error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert to db")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_OutboxCreateError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := sampleExpenseCategory()
	wallet := sampleWalletProto(walletTestID, 200000)
	updatedWallet := sampleWalletProto(walletTestID, 150000)
	createdTxn := sampleTransactionModel()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.Anything).Return(createdTxn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(errors.New("outbox error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateTransaction_CommitError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := sampleTransactionRequest()
	cat := sampleExpenseCategory()
	wallet := sampleWalletProto(walletTestID, 200000)
	updatedWallet := sampleWalletProto(walletTestID, 150000)
	createdTxn := sampleTransactionModel()

	d.categoryRepo.On("GetCategoryByID", mock.Anything, mock.Anything, catTestID.String()).Return(cat, nil)
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.Anything).Return(createdTxn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(errors.New("commit error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.CreateTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

// =====================================================================
// FundTransfer
// =====================================================================

func TestFundTransfer_Success(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		CashInCategoryID:  cashInCatID.String(),
		CashOutCategoryID: cashOutCatID.String(),
		FromWalletID:      walletTestID.String(),
		ToWalletID:        wallet2ID.String(),
		Amount:            100000,
		AdminFee:          2000,
		Date:              txnFixTime,
		Description:       "Transfer dana",
	}

	fromWallet := sampleWalletProto(walletTestID, 500000)
	toWallet := sampleWalletProto(wallet2ID, 50000)
	toWallet.Id = wallet2ID.String()
	toWallet.Name = "Mandiri"

	cashOutTxn := model.Transactions{
		Base:     model.Base{ID: uuid.MustParse("55555555-5555-5555-5555-555555555555")},
		WalletID: walletTestID, CategoryID: cashOutCatID, Amount: 102000,
	}
	cashInTxn := model.Transactions{
		Base:     model.Base{ID: uuid.MustParse("66666666-6666-6666-6666-666666666666")},
		WalletID: wallet2ID, CategoryID: cashInCatID, Amount: 100000,
	}

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(fromWallet, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, wallet2ID.String()).Return(toWallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(fromWallet, nil).Once()
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(toWallet, nil).Once()
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.MatchedBy(func(t model.Transactions) bool {
		return t.WalletID == walletTestID // cash out
	})).Return(cashOutTxn, nil)
	d.transactionRepo.On("CreateTransaction", mock.Anything, d.tx, mock.MatchedBy(func(t model.Transactions) bool {
		return t.WalletID == wallet2ID // cash in
	})).Return(cashInTxn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil).Times(2)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.FundTransfer(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.CashOutTransactionID)
	assert.NotEmpty(t, result.CashInTransactionID)
	assert.Equal(t, walletTestID.String(), result.FromWalletID)
	assert.Equal(t, wallet2ID.String(), result.ToWalletID)
	assert.Equal(t, float64(100000), result.Amount)
	d.assertAll(t)
}

func TestFundTransfer_BeginTxError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		FromWalletID: walletTestID.String(),
		ToWalletID:   wallet2ID.String(),
		Amount:       100000,
	}
	d.txManager.On("Begin", mock.Anything).Return(nil, errors.New("begin error"))

	result, err := svc.FundTransfer(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
	assert.Empty(t, result.FromWalletID)
	d.assertAll(t)
}

func TestFundTransfer_FromWalletNotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		FromWalletID: walletTestID.String(),
		ToWalletID:   wallet2ID.String(),
		Amount:       100000,
	}
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).
		Return(nil, errors.New("wallet not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.FundTransfer(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source wallet not found")
	assert.Empty(t, result.FromWalletID)
	d.assertAll(t)
}

func TestFundTransfer_ToWalletNotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		FromWalletID: walletTestID.String(),
		ToWalletID:   wallet2ID.String(),
		Amount:       100000,
	}
	fromWallet := sampleWalletProto(walletTestID, 500000)

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(fromWallet, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, wallet2ID.String()).
		Return(nil, errors.New("wallet not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.FundTransfer(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination wallet not found")
	assert.Empty(t, result.FromWalletID)
	d.assertAll(t)
}

func TestFundTransfer_InsufficientBalance(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		FromWalletID: walletTestID.String(),
		ToWalletID:   wallet2ID.String(),
		Amount:       900000,
		AdminFee:     2000,
	}
	fromWallet := sampleWalletProto(walletTestID, 100000) // less than amount+fee
	toWallet := sampleWalletProto(wallet2ID, 50000)
	toWallet.Id = wallet2ID.String()

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(fromWallet, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, wallet2ID.String()).Return(toWallet, nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.FundTransfer(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient wallet balance")
	assert.Empty(t, result.FromWalletID)
	d.assertAll(t)
}

func TestFundTransfer_SameWallet(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		FromWalletID: walletTestID.String(),
		ToWalletID:   walletTestID.String(), // same wallet
		Amount:       100000,
	}
	wallet := sampleWalletProto(walletTestID, 500000)

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil).Times(2)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.FundTransfer(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be the same")
	assert.Empty(t, result.FromWalletID)
	d.assertAll(t)
}

func TestFundTransfer_InvalidFromCategoryID(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	req := dto.FundTransferRequest{
		CashOutCategoryID: "not-a-uuid",
		CashInCategoryID:  cashInCatID.String(),
		FromWalletID:      walletTestID.String(),
		ToWalletID:        wallet2ID.String(),
		Amount:            100000,
	}
	fromWallet := sampleWalletProto(walletTestID, 500000)
	toWallet := sampleWalletProto(wallet2ID, 50000)
	toWallet.Id = wallet2ID.String()

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(fromWallet, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, wallet2ID.String()).Return(toWallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(fromWallet, nil).Maybe()
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(toWallet, nil).Maybe()
	d.tx.On("Rollback").Return(nil)

	result, err := svc.FundTransfer(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid from category id")
	assert.Empty(t, result.FromWalletID)
	d.assertAll(t)
}

// =====================================================================
// DeleteTransaction
// =====================================================================

func TestDeleteTransaction_SuccessExpense(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel() // expense type
	wallet := sampleWalletProto(walletTestID, 50000)
	updatedWallet := sampleWalletProto(walletTestID, 100000) // balance restored

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(txn, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("DeleteTransaction", mock.Anything, d.tx, txn).Return(txn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), txnTestID.String())

	assert.NoError(t, err)
	assert.Equal(t, txnTestID.String(), result.ID)
	d.assertAll(t)
}

func TestDeleteTransaction_SuccessIncome(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	txn.Category = sampleIncomeCategory()
	wallet := sampleWalletProto(walletTestID, 500000)
	updatedWallet := sampleWalletProto(walletTestID, 450000)

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(txn, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("DeleteTransaction", mock.Anything, d.tx, txn).Return(txn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), txnTestID.String())

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	d.assertAll(t)
}

func TestDeleteTransaction_SuccessFundTransferCashOut(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	txn.Category = sampleFundTransferCashOut()
	wallet := sampleWalletProto(walletTestID, 100000)
	updatedWallet := sampleWalletProto(walletTestID, 150000) // balance restored after cancelling cash out

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(txn, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("DeleteTransaction", mock.Anything, d.tx, txn).Return(txn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(nil)
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), txnTestID.String())

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	d.assertAll(t)
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, "bad-id").
		Return(model.Transactions{}, errors.New("transaction not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), "bad-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestDeleteTransaction_WalletNotFound(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(txn, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).
		Return(nil, errors.New("wallet not found"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), txnTestID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wallet not found")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestDeleteTransaction_DeleteDBError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	wallet := sampleWalletProto(walletTestID, 50000)
	updatedWallet := sampleWalletProto(walletTestID, 100000)

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(txn, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("DeleteTransaction", mock.Anything, d.tx, txn).
		Return(model.Transactions{}, errors.New("db delete error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), txnTestID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete from db")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestDeleteTransaction_CommitError(t *testing.T) {
	d := newTransactionTestDeps()
	svc := d.service()

	txn := sampleTransactionModel()
	wallet := sampleWalletProto(walletTestID, 50000)
	updatedWallet := sampleWalletProto(walletTestID, 100000)

	d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
	d.transactionRepo.On("GetTransactionByID", mock.Anything, d.tx, txnTestID.String()).Return(txn, nil)
	d.walletClient.On("GetWalletByID", mock.Anything, walletTestID.String()).Return(wallet, nil)
	d.walletClient.On("UpdateWallet", mock.Anything, mock.Anything).Return(updatedWallet, nil)
	d.transactionRepo.On("DeleteTransaction", mock.Anything, d.tx, txn).Return(txn, nil)
	d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
	d.tx.On("Commit").Return(errors.New("commit error"))
	d.tx.On("Rollback").Return(nil)

	result, err := svc.DeleteTransaction(context.Background(), txnTestID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit")
	assert.Empty(t, result.ID)
	d.assertAll(t)
}
