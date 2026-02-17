package model

import "time"

type OutboxMessage struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	AggregateID string     `gorm:"type:uuid;index;not null" json:"aggregate_id"`
	EventType   string     `gorm:"index;not null" json:"event_type"`
	Payload     []byte     `gorm:"type:jsonb;not null" json:"payload"`
	Published   bool       `gorm:"index;default:false" json:"published"`
	PublishedAt *time.Time `json:"published_at"`
	Retries     int        `gorm:"default:0" json:"retries"`
	MaxRetries  int        `gorm:"default:3" json:"max_retries"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
