package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringSlice is a custom type for handling string arrays in JSON
type StringSlice []string

// Value implements the driver.Valuer interface
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, s)
}

// GoogleAccount represents a connected Google account
type GoogleAccount struct {
	gorm.Model
	UserID         uint        `json:"user_id" gorm:"not null"`
	GoogleID       string      `json:"google_id" gorm:"unique;not null"`
	Email          string      `json:"email" gorm:"unique;not null"`
	AccessToken    string      `json:"-" gorm:"not null"` // OAuth access token
	RefreshToken   string      `json:"-" gorm:"not null"` // OAuth refresh token
	TokenExpiry    time.Time   `json:"token_expiry" gorm:"not null"`
	CalendarIDs    StringSlice `json:"calendar_ids" gorm:"type:json"`
	IsActive       bool        `json:"is_active" gorm:"default:true"`
	LastSyncAt     time.Time   `json:"last_sync_at"`
	ProfilePicture string      `json:"profile_picture" gorm:"type:text"`
	Name           string      `json:"name"`
	
	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the GoogleAccount model
func (GoogleAccount) TableName() string {
	return "google_accounts"
} 