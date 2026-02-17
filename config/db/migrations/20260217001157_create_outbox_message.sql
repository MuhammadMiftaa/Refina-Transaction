-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox_messages (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP,
    retries INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_outbox_published ON outbox_messages(published);
CREATE INDEX idx_outbox_aggregate_id ON outbox_messages(aggregate_id);
CREATE INDEX idx_outbox_event_type ON outbox_messages(event_type);
CREATE INDEX idx_outbox_created_at ON outbox_messages(created_at);

-- Index for cleanup job
CREATE INDEX idx_outbox_published_at ON outbox_messages(published_at) WHERE published = TRUE;

-- Composite index for query pending messages
CREATE INDEX idx_outbox_pending ON outbox_messages(published, retries, created_at) WHERE published = FALSE;

COMMENT ON TABLE outbox_messages IS 'Outbox pattern table for reliable event publishing in CQRS';
COMMENT ON COLUMN outbox_messages.aggregate_id IS 'ID of the aggregate (e.g., wallet_id)';
COMMENT ON COLUMN outbox_messages.event_type IS 'Type of event (e.g., wallet.created, wallet.updated)';
COMMENT ON COLUMN outbox_messages.payload IS 'Event payload in JSON format';
COMMENT ON COLUMN outbox_messages.published IS 'Whether the message has been published to message broker';
COMMENT ON COLUMN outbox_messages.retries IS 'Number of retry attempts';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_outbox_pending;
DROP INDEX IF EXISTS idx_outbox_published_at;
DROP INDEX IF EXISTS idx_outbox_created_at;
DROP INDEX IF EXISTS idx_outbox_event_type;
DROP INDEX IF EXISTS idx_outbox_aggregate_id;
DROP INDEX IF EXISTS idx_outbox_published;

DROP TABLE IF EXISTS outbox_messages;
-- +goose StatementEnd