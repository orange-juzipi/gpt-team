package model

import (
	"time"

	"gorm.io/gorm"
)

type CardStatus string
type CardType string

const (
	CardStatusUnactivated CardStatus = "unactivated"
	CardStatusActivated   CardStatus = "activated"

	CardTypeUS CardType = "us"
	CardTypeUK CardType = "uk"
	CardTypeES CardType = "es"
)

type Card struct {
	ID           uint64     `gorm:"primaryKey"`
	Code         string     `gorm:"size:128;uniqueIndex;not null"`
	CardType     CardType   `gorm:"size:16;not null;default:us"`
	CardLimit    int        `gorm:"not null;default:0"`
	Status       CardStatus `gorm:"size:32;not null;default:unactivated"`
	RemoteStatus string     `gorm:"size:64"`
	RemoteCardID *uint64    `gorm:"index"`
	LastFour     string     `gorm:"size:4"`
	ExpiryDate   string     `gorm:"size:16"`
	FullName     string     `gorm:"size:255"`
	Birthday     string     `gorm:"size:64"`
	LastSyncedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
