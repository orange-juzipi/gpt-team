package model

import (
	"time"

	"gorm.io/gorm"
)

type AccountType string
type AccountStatus string

const (
	AccountTypePlus     AccountType = "plus"
	AccountTypeBusiness AccountType = "business"

	AccountStatusNormal  AccountStatus = "normal"
	AccountStatusBlocked AccountStatus = "blocked"
)

type Account struct {
	ID                 uint64      `gorm:"primaryKey"`
	Account            string      `gorm:"size:255;not null"`
	PasswordCiphertext string      `gorm:"type:text;not null"`
	Type               AccountType `gorm:"size:32;not null"`
	StartTime          *time.Time
	EndTime            *time.Time
	Status             AccountStatus `gorm:"size:32;not null;default:normal"`
	Remark             string        `gorm:"type:text"`
	ParentID           *uint64       `gorm:"index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}
