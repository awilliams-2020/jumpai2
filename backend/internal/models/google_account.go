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
	UserID         uint        `gorm:"not null" json:"user_id"`
	GoogleID       string      `gorm:"not null" json:"google_id"`
	Email          string      `gorm:"not null" json:"email"`
	AccessToken    string      `gorm:"not null" json:"access_token"`
	RefreshToken   string      `gorm:"not null" json:"refresh_token"`
	TokenExpiry    time.Time   `gorm:"not null" json:"token_expiry"`
	CalendarIDs    StringSlice `gorm:"type:json" json:"calendar_ids"`
	IsActive       bool        `gorm:"default:true" json:"is_active"`
	LastSyncedAt   time.Time   `json:"last_synced_at"`
	ProfilePicture string      `json:"profile_picture"`
	Name           string      `json:"name"`
} 