package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"refina-transaction/internal/service/mocks"
	"refina-transaction/internal/types/model"
	"refina-transaction/internal/utils/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// Test Dependency Container
// ─────────────────────────────────────────────

type outboxTestDeps struct {
	outboxRepo *mocks.MockOutboxRepository
	queue      *mocks.MockRabbitMQClient
}

func newOutboxTestDeps() *outboxTestDeps {
	return &outboxTestDeps{
		outboxRepo: new(mocks.MockOutboxRepository),
		queue:      new(mocks.MockRabbitMQClient),
	}
}

func (d *outboxTestDeps) publisher() *OutboxPublisher {
	return NewOutboxPublisher(d.outboxRepo, d.queue)
}

func (d *outboxTestDeps) assertAll(t *testing.T) {
	t.Helper()
	d.outboxRepo.AssertExpectations(t)
	d.queue.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Sample Data Factories
// ─────────────────────────────────────────────

func sampleOutboxMessage() model.OutboxMessage {
	return model.OutboxMessage{
		ID:          1,
		AggregateID: txnTestID.String(),
		EventType:   data.OUTBOX_EVENT_TRANSACTION_CREATED,
		Payload:     []byte(`{"id":"test"}`),
		Published:   false,
		Retries:     0,
		MaxRetries:  5,
		CreatedAt:   time.Now(),
	}
}

// =====================================================================
// NewOutboxPublisher
// =====================================================================

func TestNewOutboxPublisher_DefaultConfig(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()

	assert.NotNil(t, pub)
	assert.Equal(t, data.OUTBOX_PUBLISH_INTERVAL, pub.interval)
	assert.Equal(t, data.OUTBOX_PUBLISH_BATCH, pub.batchSize)
}

// =====================================================================
// Start — context cancellation
// =====================================================================

func TestStart_ContextCancellation(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()
	pub.interval = 10 * time.Millisecond // speed up ticker

	// GetPendingMessages may be called 0 or more times during the brief window
	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// goroutine exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("Start goroutine did not stop after context cancellation")
	}
}

// =====================================================================
// publishPendingMessages (tested via Start with short interval)
// =====================================================================

func TestPublishPendingMessages_NoPendingMessages(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()
	pub.interval = 10 * time.Millisecond

	// Only return empty list — no publish should happen
	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.Start(ctx)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	<-done
	d.queue.AssertNotCalled(t, "GetChannel")
}

func TestPublishPendingMessages_GetPendingMessagesError(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()
	pub.interval = 10 * time.Millisecond

	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{}, errors.New("db error")).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.Start(ctx)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	<-done
	// Should log error but not crash — goroutine stopped cleanly
	d.queue.AssertNotCalled(t, "GetChannel")
}

// =====================================================================
// publishMessage — channel error handling
// =====================================================================

func TestPublishMessage_GetChannelError(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()
	pub.interval = 10 * time.Millisecond

	msg := sampleOutboxMessage()

	// GetPendingMessages returns one message
	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{msg}, nil).Once()
	// Subsequent calls return empty
	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{}, nil).Maybe()

	// GetChannel fails → publishMessage fails → IncrementRetries called
	d.queue.On("GetChannel").Return(nil, errors.New("channel error")).Once()
	d.outboxRepo.On("IncrementRetries", mock.Anything, uint(1)).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	<-done
	d.assertAll(t)
}

func TestPublishMessage_MaxRetriesExceeded(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()
	pub.interval = 10 * time.Millisecond

	// Message that has already hit max retries
	msg := sampleOutboxMessage()
	msg.Retries = msg.MaxRetries - 1 // on the edge

	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{msg}, nil).Once()
	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{}, nil).Maybe()

	d.queue.On("GetChannel").Return(nil, errors.New("channel error")).Once()
	d.outboxRepo.On("IncrementRetries", mock.Anything, uint(1)).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	<-done
	// Max retries log is emitted — just verify mocks were called
	d.assertAll(t)
}

func TestPublishMessage_IncrementRetriesError(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()
	pub.interval = 10 * time.Millisecond

	msg := sampleOutboxMessage()

	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{msg}, nil).Once()
	d.outboxRepo.On("GetPendingMessages", mock.Anything, pub.batchSize).
		Return([]model.OutboxMessage{}, nil).Maybe()

	d.queue.On("GetChannel").Return(nil, errors.New("channel error")).Once()
	// IncrementRetries itself fails — should log but not crash
	d.outboxRepo.On("IncrementRetries", mock.Anything, uint(1)).Return(errors.New("db error")).Once()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	<-done
	d.assertAll(t)
}

func TestPublishMessage_MarkAsPublishedError(t *testing.T) {
	// publishMessage success but MarkAsPublished fails
	// This tests the channel path which requires a real amqp channel.
	// Covered by integration tests.
	t.Skip("publishMessage requires a real amqp091.Channel; covered by integration tests")
}

// =====================================================================
// StartCleanupJob — context cancellation
// =====================================================================

func TestStartCleanupJob_ContextCancellation(t *testing.T) {
	d := newOutboxTestDeps()
	pub := d.publisher()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pub.StartCleanupJob(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// stopped cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("StartCleanupJob did not stop after context cancellation")
	}
}
