package model

import (
	"time"

	"github.com/google/uuid"
)

type Transactions struct {
	Base
	WalletID        uuid.UUID `gorm:"type:uuid;not null"`
	CategoryID      uuid.UUID `gorm:"type:uuid;not null"`
	Amount          float64   `gorm:"type:decimal(18,2);not null"`
	TransactionDate time.Time `gorm:"type:timestamp;not null"`
	Description     string    `gorm:"type:text"`

	Category    Categories    `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Attachments []Attachments `gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
