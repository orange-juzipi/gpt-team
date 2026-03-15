package model

import (
	"time"

	"gorm.io/gorm"
)

type MailboxProvider struct {
	ID              uint64              `gorm:"primaryKey"`
	ProviderType    MailboxProviderType `gorm:"size:32;not null;default:'cloudmail'"`
	DomainSuffix    string              `gorm:"size:255;uniqueIndex;not null"`
	AccountEmail    string              `gorm:"size:255;not null;default:''"`
	AccountID       *uint64
	TokenCiphertext string `gorm:"type:text;not null"`
	Remark          string `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type MailboxProviderType string

const (
	MailboxProviderTypeCloudmail MailboxProviderType = "cloudmail"
	MailboxProviderTypeDuckmail  MailboxProviderType = "duckmail"
)
