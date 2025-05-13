package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model
	Email          string    `json:"email" gorm:"unique;not null"`
	Name           string    `json:"name" gorm:"not null"`
	ProfilePicture string    `json:"profile_picture" gorm:"type:text"`
	GoogleID       string    `json:"google_id" gorm:"unique"`
	HubspotID      *string   `json:"hubspot_id" gorm:"unique"`
	AccessToken    string    `json:"-" gorm:"type:text"`
	RefreshToken   string    `json:"-" gorm:"type:text"`
	TokenExpiry    time.Time `json:"-"`
	CalendarIDs    []string  `json:"calendar_ids" gorm:"type:json"`
	LastLoginAt    time.Time `json:"last_login_at"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	
	// Relationships
	GoogleAccounts    []GoogleAccount    `json:"google_accounts" gorm:"foreignKey:UserID"`
	HubspotAccounts   []HubSpotAccount   `json:"hubspot_accounts" gorm:"foreignKey:UserID"`
	SchedulingWindows []SchedulingWindow `json:"scheduling_windows" gorm:"foreignKey:UserID"`
	SchedulingLinks   []SchedulingLink   `json:"scheduling_links" gorm:"foreignKey:UserID"`
	Meetings          []Meeting          `json:"meetings" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

type SchedulingWindow struct {
	gorm.Model
	UserID        uint      `gorm:"not null"`
	StartHour     int       `gorm:"not null"`
	EndHour       int       `gorm:"not null"`
	Weekday       int       `gorm:"not null"` // 0-6 (Sunday-Saturday)
	IsActive      bool      `gorm:"default:true"`
}

type SchedulingLink struct {
	gorm.Model
	UserID            uint      `gorm:"not null"`
	Title             string    `gorm:"not null"`
	Duration          int       `gorm:"not null"` // in minutes
	MaxUses           *int      // nil means unlimited
	ExpiresAt         *time.Time
	MaxDaysInAdvance  int       `gorm:"not null"`
	CustomQuestions   string    `gorm:"type:json"` // Store as JSON string
	IsActive          bool      `gorm:"default:true"`
}

type Meeting struct {
	gorm.Model
	SchedulingLinkID  uint      `gorm:"not null"`
	UserID            uint      `gorm:"not null"`
	ClientEmail       string    `gorm:"not null"`
	LinkedInURL       string
	StartTime         time.Time `gorm:"not null"`
	EndTime           time.Time `gorm:"not null"`
	Answers           StringSlice `gorm:"type:json"`
	HubspotContactID  *string
	LinkedInData      string    `gorm:"type:json"`
	ContextNotes      string    `gorm:"type:text"`
}