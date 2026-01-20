package model

import "github.com/google/uuid"

type CategoryType string

const (
	Income       CategoryType = "income"
	Expense      CategoryType = "expense"
	FundTransfer CategoryType = "fund_transfer"
)

type Categories struct {
	Base
	ParentID *uuid.UUID   `gorm:"type:uuid"`
	Name     string       `gorm:"type:varchar(50);not null"`
	Type     CategoryType `gorm:"type:varchar(50);not null"`

	Parent   *Categories  `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`
	Children []Categories `gorm:"foreignKey:ParentID"`
}
