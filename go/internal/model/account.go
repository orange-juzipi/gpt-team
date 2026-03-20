package model

import (
	"time"

	"gorm.io/gorm"
)

type AccountType string
type AccountStatus string
type AccountRelationType string

const (
	AccountTypePlus     AccountType = "plus"
	AccountTypeBusiness AccountType = "business"
	AccountTypeCodex    AccountType = "codex"

	AccountStatusNormal  AccountStatus = "normal"
	AccountStatusBlocked AccountStatus = "blocked"

	AccountRelationTypeWarranty   AccountRelationType = "warranty"
	AccountRelationTypeSubAccount AccountRelationType = "subaccount"
)

type Account struct {
	ID                 uint64      `gorm:"primaryKey"`
	Account            string      `gorm:"size:255;not null"`
	PasswordCiphertext string      `gorm:"type:text;not null"`
	Type               AccountType `gorm:"size:32;not null"`
	StartTime          *time.Time
	EndTime            *time.Time
	Status             AccountStatus       `gorm:"size:32;not null;default:normal"`
	Remark             string              `gorm:"type:text"`
	ParentID           *uint64             `gorm:"index"`
	RelationType       AccountRelationType `gorm:"size:32"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}
