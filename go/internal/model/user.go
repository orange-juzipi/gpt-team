package model

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
)

type User struct {
	ID           uint64   `gorm:"primaryKey"`
	Username     string   `gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string   `gorm:"type:text;not null"`
	Role         UserRole `gorm:"size:32;not null;default:member"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
