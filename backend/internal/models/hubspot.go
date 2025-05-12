package models

import (
	"time"

	"gorm.io/gorm"
)

// HubSpotAccount represents a connected HubSpot account
type HubSpotAccount struct {
	gorm.Model
	UserID         uint      `json:"user_id" gorm:"not null"`
	HubID          string    `json:"hub_id" gorm:"uniqueIndex;not null"`
	HubName        string    `json:"hub_name"`
	HubDomain      string    `json:"hub_domain"`
	AccessToken    string    `json:"-" gorm:"not null"` // OAuth access token
	RefreshToken   string    `json:"-" gorm:"not null"` // OAuth refresh token
	TokenExpiry    time.Time `json:"token_expiry"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	LastSyncedAt   time.Time `json:"last_synced_at"`
	HubTimezone    string    `json:"hub_timezone"`
	User           User      `json:"user" gorm:"foreignKey:UserID"`
} 