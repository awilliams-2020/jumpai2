package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"gorm.io/gorm"
)

type GoogleHandler struct {
	db *gorm.DB
}

func NewGoogleHandler(db *gorm.DB) *GoogleHandler {
	return &GoogleHandler{
		db: db,
	}
}

// Profile returns the user's profile information
func (h *GoogleHandler) Profile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Type assert the user
	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             userModel.ID,
		"email":          userModel.Email,
		"name":           userModel.Name,
		"profile_picture": userModel.ProfilePicture,
		"created_at":     userModel.CreatedAt,
		"updated_at":     userModel.UpdatedAt,
		"is_active":      userModel.IsActive,
		"last_login_at":  userModel.LastLoginAt,
	})
}

// GetGoogleAccounts retrieves all connected Google accounts for the authenticated user
func (h *GoogleHandler) GetGoogleAccounts(c *gin.Context) {
	userID := c.GetUint("user_id")

	var accounts []models.GoogleAccount
	if err := h.db.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Google accounts"})
		return
	}

	// Format response to ensure snake_case
	response := make([]gin.H, len(accounts))
	for i, account := range accounts {
		response[i] = gin.H{
			"id":              account.ID,
			"user_id":         account.UserID,
			"google_id":       account.GoogleID,
			"email":           account.Email,
			"name":            account.Name,
			"profile_picture": account.ProfilePicture,
			"is_active":       account.IsActive,
			"last_synced_at":  account.LastSyncedAt,
			"calendar_ids":    account.CalendarIDs,
		}
	}

	c.JSON(http.StatusOK, response)
}

// DisconnectGoogleAccount disconnects a Google account
func (h *GoogleHandler) DisconnectGoogleAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	accountID := c.Param("id")

	var account models.GoogleAccount
	if err := h.db.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Google account not found"})
		return
	}

	if err := h.db.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Google account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Google account disconnected successfully"})
} 