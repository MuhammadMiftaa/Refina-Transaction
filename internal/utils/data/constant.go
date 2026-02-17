package data

import "time"

var (
	DEVELOPMENT_MODE = "development"
	STAGING_MODE     = "staging"
	PRODUCTION_MODE  = "production"

	OUTBOX_PUBLISH_EXCHANGE          = "refina_microservice"
	OUTBOX_PUBLISH_INTERVAL          = 5 * time.Second
	OUTBOX_PUBLISH_BATCH             = 100
	OUTBOX_PUBLISH_MAX_RETRIES       = 5
	OUTBOX_EVENT_TRANSACTION_CREATED = "transaction.created"
	OUTBOX_EVENT_TRANSACTION_UPDATED = "transaction.updated"
	OUTBOX_EVENT_TRANSACTION_DELETED = "transaction.deleted"
)
