package models

import (
	"time"

	"gorm.io/gorm"
)

// HubSpotAccount represents a connected HubSpot account
type HubSpotAccount struct {
	gorm.Model
	UserID       uint      `json:"user_id" gorm:"not null"`
	HubID        string    `json:"hub_id" gorm:"uniqueIndex;not null"`
	HubName      string    `json:"hub_name" gorm:"not null"`
	HubDomain    string    `json:"hub_domain" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	AccessToken  string    `json:"-" gorm:"not null"` // OAuth access token
	RefreshToken string    `json:"-" gorm:"not null"` // OAuth refresh token
	TokenExpiry  time.Time `json:"token_expiry" gorm:"not null"`
	HubTimezone  string    `json:"hub_timezone" gorm:"not null"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the HubSpotAccount model
func (HubSpotAccount) TableName() string {
	return "hubspot_accounts"
} 