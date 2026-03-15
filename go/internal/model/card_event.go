package model

import "time"

type CardEventType string

const (
	CardEventActivate        CardEventType = "activate"
	CardEventQuery           CardEventType = "query"
	CardEventBilling         CardEventType = "billing"
	CardEventThreeDS         CardEventType = "three_ds"
	CardEventIdentityRefresh CardEventType = "identity_refresh"
)

type CardEvent struct {
	ID                uint64        `gorm:"primaryKey"`
	CardID            uint64        `gorm:"index;not null"`
	EventType         CardEventType `gorm:"size:32;index;not null"`
	RequestPayload    string        `gorm:"type:text;not null"`
	ResponsePayload   string        `gorm:"type:text;not null"`
	NormalizedPayload string        `gorm:"type:text;not null"`
	Success           bool          `gorm:"not null"`
	ErrorMessage      string        `gorm:"type:text"`
	CreatedAt         time.Time
}
