package model

import "github.com/google/uuid"

type Attachments struct {
	Base
	TransactionID uuid.UUID `gorm:"type:uuid;not null"`
	Image         string    `gorm:"type:text"`
	Format        string    `gorm:"type:text"`
	Size          int64     `gorm:"type:bigint"`

	Transaction Transactions `gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
