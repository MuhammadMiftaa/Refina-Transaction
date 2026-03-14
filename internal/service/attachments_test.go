package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"refina-transaction/internal/service/mocks"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// Test Dependency Container
// ─────────────────────────────────────────────

type attachmentTestDeps struct {
	txManager      *mocks.MockTxManager
	attachmentRepo *mocks.MockAttachmentsRepository
	tx             *mocks.MockTransaction
}

func newAttachmentTestDeps() *attachmentTestDeps {
	return &attachmentTestDeps{
		txManager:      new(mocks.MockTxManager),
		attachmentRepo: new(mocks.MockAttachmentsRepository),
		tx:             new(mocks.MockTransaction),
	}
}

func (d *attachmentTestDeps) service() AttachmentsService {
	return NewAttachmentsService(d.txManager, d.attachmentRepo)
}

func (d *attachmentTestDeps) assertAll(t *testing.T) {
	t.Helper()
	d.txManager.AssertExpectations(t)
	d.attachmentRepo.AssertExpectations(t)
	d.tx.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Fixed UUIDs & Timestamps
// ─────────────────────────────────────────────

var (
	attID   = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	txnID   = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	attTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
)

// ─────────────────────────────────────────────
// Sample Data Factories
// ─────────────────────────────────────────────

func sampleAttachmentModel() model.Attachments {
	return model.Attachments{
		Base:          model.Base{ID: attID, CreatedAt: attTime, UpdatedAt: attTime},
		TransactionID: txnID,
		Image:         "http://minio/bucket/image.jpg",
		Format:        ".jpg",
		Size:          1024,
	}
}

// =====================================================================
// GetAllAttachments
// =====================================================================

func TestGetAllAttachments_Success(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	attachments := []model.Attachments{sampleAttachmentModel()}
	d.attachmentRepo.On("GetAllAttachments", mock.Anything, nil).Return(attachments, nil)

	result, err := svc.GetAllAttachments(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, attID.String(), result[0].ID)
	assert.Equal(t, txnID.String(), result[0].TransactionID)
	assert.Equal(t, "http://minio/bucket/image.jpg", result[0].Image)
	d.assertAll(t)
}

func TestGetAllAttachments_EmptyList(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAllAttachments", mock.Anything, nil).Return([]model.Attachments{}, nil)

	result, err := svc.GetAllAttachments(context.Background())

	assert.NoError(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

func TestGetAllAttachments_RepositoryError(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAllAttachments", mock.Anything, nil).
		Return([]model.Attachments{}, errors.New("db error"))

	result, err := svc.GetAllAttachments(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

// =====================================================================
// GetAttachmentByID
// =====================================================================

func TestGetAttachmentByID_Success(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	att := sampleAttachmentModel()
	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, attID.String()).Return(att, nil)

	result, err := svc.GetAttachmentByID(context.Background(), attID.String())

	assert.NoError(t, err)
	assert.Equal(t, attID.String(), result.ID)
	assert.Equal(t, txnID.String(), result.TransactionID)
	d.assertAll(t)
}

func TestGetAttachmentByID_NotFound(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, "bad-id").
		Return(model.Attachments{}, errors.New("record not found"))

	result, err := svc.GetAttachmentByID(context.Background(), "bad-id")

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

// =====================================================================
// GetAttachmentsByTransactionID
// =====================================================================

func TestGetAttachmentsByTransactionID_Success(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	attachments := []model.Attachments{sampleAttachmentModel()}
	d.attachmentRepo.On("GetAttachmentsByTransactionID", mock.Anything, nil, txnID.String()).
		Return(attachments, nil)

	result, err := svc.GetAttachmentsByTransactionID(context.Background(), txnID.String())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, txnID.String(), result[0].TransactionID)
	d.assertAll(t)
}

func TestGetAttachmentsByTransactionID_EmptyList(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAttachmentsByTransactionID", mock.Anything, nil, txnID.String()).
		Return([]model.Attachments{}, nil)

	result, err := svc.GetAttachmentsByTransactionID(context.Background(), txnID.String())

	assert.NoError(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

func TestGetAttachmentsByTransactionID_RepositoryError(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAttachmentsByTransactionID", mock.Anything, nil, txnID.String()).
		Return([]model.Attachments{}, errors.New("db error"))

	result, err := svc.GetAttachmentsByTransactionID(context.Background(), txnID.String())

	assert.Error(t, err)
	assert.Nil(t, result)
	d.assertAll(t)
}

// =====================================================================
// CreateAttachment
// =====================================================================

func TestCreateAttachment_Success(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	att := sampleAttachmentModel()
	req := dto.AttachmentsRequest{
		TransactionID: txnID.String(),
		Image:         "http://minio/bucket/image.jpg",
	}

	d.attachmentRepo.On("CreateAttachment", mock.Anything, nil, mock.MatchedBy(func(a model.Attachments) bool {
		return a.TransactionID == txnID && a.Image == req.Image
	})).Return(att, nil)

	result, err := svc.CreateAttachment(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, attID.String(), result.ID)
	assert.Equal(t, txnID.String(), result.TransactionID)
	d.assertAll(t)
}

func TestCreateAttachment_InvalidTransactionID(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	req := dto.AttachmentsRequest{
		TransactionID: "not-a-uuid",
		Image:         "http://minio/bucket/image.jpg",
	}

	result, err := svc.CreateAttachment(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestCreateAttachment_RepositoryError(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	req := dto.AttachmentsRequest{
		TransactionID: txnID.String(),
		Image:         "http://minio/bucket/image.jpg",
	}

	d.attachmentRepo.On("CreateAttachment", mock.Anything, nil, mock.Anything).
		Return(model.Attachments{}, errors.New("db error"))

	result, err := svc.CreateAttachment(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

// =====================================================================
// UpdateAttachment
// =====================================================================

func TestUpdateAttachment_Success(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	existing := sampleAttachmentModel()
	updated := existing
	updated.Image = "http://minio/bucket/new_image.png"

	req := dto.AttachmentsRequest{Image: "http://minio/bucket/new_image.png"}

	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, attID.String()).Return(existing, nil)
	d.attachmentRepo.On("UpdateAttachment", mock.Anything, nil, mock.MatchedBy(func(a model.Attachments) bool {
		return a.Image == "http://minio/bucket/new_image.png"
	})).Return(updated, nil)

	result, err := svc.UpdateAttachment(context.Background(), attID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, attID.String(), result.ID)
	assert.Equal(t, "http://minio/bucket/new_image.png", result.Image)
	d.assertAll(t)
}

func TestUpdateAttachment_NoImageChange(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	existing := sampleAttachmentModel()
	req := dto.AttachmentsRequest{Image: ""} // empty means no change

	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, attID.String()).Return(existing, nil)
	d.attachmentRepo.On("UpdateAttachment", mock.Anything, nil, mock.MatchedBy(func(a model.Attachments) bool {
		// image should remain unchanged
		return a.Image == existing.Image
	})).Return(existing, nil)

	result, err := svc.UpdateAttachment(context.Background(), attID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, existing.Image, result.Image)
	d.assertAll(t)
}

func TestUpdateAttachment_NotFound(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, "bad-id").
		Return(model.Attachments{}, errors.New("record not found"))

	result, err := svc.UpdateAttachment(context.Background(), "bad-id", dto.AttachmentsRequest{Image: "x"})

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestUpdateAttachment_RepositoryUpdateError(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	existing := sampleAttachmentModel()
	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, attID.String()).Return(existing, nil)
	d.attachmentRepo.On("UpdateAttachment", mock.Anything, nil, mock.Anything).
		Return(model.Attachments{}, errors.New("db error"))

	result, err := svc.UpdateAttachment(context.Background(), attID.String(), dto.AttachmentsRequest{Image: "new"})

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

// =====================================================================
// DeleteAttachment
// =====================================================================

func TestDeleteAttachment_Success(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	att := sampleAttachmentModel()
	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, attID.String()).Return(att, nil)
	d.attachmentRepo.On("DeleteAttachment", mock.Anything, nil, att).Return(att, nil)

	result, err := svc.DeleteAttachment(context.Background(), attID.String())

	assert.NoError(t, err)
	assert.Equal(t, attID.String(), result.ID)
	d.assertAll(t)
}

func TestDeleteAttachment_NotFound(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, "bad-id").
		Return(model.Attachments{}, errors.New("record not found"))

	result, err := svc.DeleteAttachment(context.Background(), "bad-id")

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}

func TestDeleteAttachment_RepositoryDeleteError(t *testing.T) {
	d := newAttachmentTestDeps()
	svc := d.service()

	att := sampleAttachmentModel()
	d.attachmentRepo.On("GetAttachmentByID", mock.Anything, nil, attID.String()).Return(att, nil)
	d.attachmentRepo.On("DeleteAttachment", mock.Anything, nil, att).
		Return(model.Attachments{}, errors.New("db error"))

	result, err := svc.DeleteAttachment(context.Background(), attID.String())

	assert.Error(t, err)
	assert.Empty(t, result.ID)
	d.assertAll(t)
}
