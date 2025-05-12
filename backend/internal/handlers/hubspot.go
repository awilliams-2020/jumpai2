package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var (
	hubspotOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("HUBSPOT_CLIENT_ID"),
		ClientSecret: os.Getenv("HUBSPOT_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("HUBSPOT_REDIRECT_URL"),
		Scopes: []string{
			"crm.objects.contacts.read",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://app.hubspot.com/oauth/authorize",
			TokenURL: "https://api.hubspot.com/oauth/v1/token",
		},
	}
)

type HubSpotHandler struct {
	db *gorm.DB
}

func NewHubSpotHandler(db *gorm.DB) *HubSpotHandler {
	return &HubSpotHandler{db: db}
}

// HubSpotConnect initiates the OAuth 2.0 flow for connecting HubSpot
func (h *HubSpotHandler) HubSpotConnect(c *gin.Context) {
	// Get token from query parameter
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token required"})
		return
	}

	// Verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token claims"})
		return
	}

	userID, ok := claims["sub"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Set the state parameter to include the user ID
	state := fmt.Sprintf("connect_%d", uint(userID))

	// Store the token in a cookie for the callback
	c.SetCookie("auth_token", tokenString, 300, "/", "", false, true)

	// Redirect to HubSpot OAuth
	url := hubspotOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HubSpotConnectCallback handles the OAuth 2.0 callback from HubSpot
func (h *HubSpotHandler) HubSpotConnectCallback(c *gin.Context) {
	// Get token from cookie
	tokenString, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token claims"})
		return
	}

	userID, ok := claims["sub"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Verify state parameter
	state := c.Query("state")
	expectedState := fmt.Sprintf("connect_%d", uint(userID))
	if state != expectedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	code := c.Query("code")
	hubspotToken, err := hubspotOAuthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info from HubSpot
	client := hubspotOAuthConfig.Client(c, hubspotToken)
	resp, err := client.Get("https://api.hubapi.com/oauth/v1/access-tokens/" + hubspotToken.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		User        string   `json:"user"`
		HubID       int64    `json:"hub_id"`
		Scopes      []string `json:"scopes"`
		ExpiresIn   int      `json:"expires_in"`
		TokenType   string   `json:"token_type"`
		AppID       int64    `json:"app_id"`
		HubDomain   string   `json:"hub_domain"`
		UserID      int64    `json:"user_id"`
		HubName     string   `json:"hub_name"`
		HubTimezone string   `json:"hub_timezone"`
		ClientID    string   `json:"client_id"`
		UserEmail   string   `json:"user_email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		// Log the error and response body for debugging
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to decode HubSpot response: %v\nResponse body: %s", err, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Get the authenticated user
	var user models.User
	if err := h.db.First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	// Check if this HubSpot account is already connected
	var existingAccount models.HubSpotAccount
	result := h.db.Where("hub_id = ?", fmt.Sprintf("%d", userInfo.HubID)).First(&existingAccount)
	if result.Error == nil {
		// Account already exists, update it
		existingAccount.AccessToken = hubspotToken.AccessToken
		existingAccount.RefreshToken = hubspotToken.RefreshToken
		existingAccount.TokenExpiry = time.Now().Add(time.Duration(userInfo.ExpiresIn) * time.Second)
		existingAccount.LastSyncAt = time.Now()
		if err := h.db.Save(&existingAccount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing account"})
			return
		}
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Create new HubSpot account
		newAccount := models.HubSpotAccount{
			UserID:         user.ID,
			HubID:          fmt.Sprintf("%d", userInfo.HubID),
			HubName:        userInfo.HubName,
			HubDomain:      userInfo.HubDomain,
			AccessToken:    hubspotToken.AccessToken,
			RefreshToken:   hubspotToken.RefreshToken,
			TokenExpiry:    time.Now().Add(time.Duration(userInfo.ExpiresIn) * time.Second),
			IsActive:       true,
			LastSyncAt:     time.Now(),
			HubTimezone:    userInfo.HubTimezone,
		}

		if err := h.db.Create(&newAccount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create HubSpot account"})
			return
		}
	}

	// Clear the auth token cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	// Redirect back to dashboard
	c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

// GetHubSpotAccounts retrieves all connected HubSpot accounts for the authenticated user
func (h *HubSpotHandler) GetHubSpotAccounts(c *gin.Context) {
	userID := c.GetUint("user_id")

	var accounts []models.HubSpotAccount
	if err := h.db.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch HubSpot accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// DisconnectAccount disconnects a HubSpot account
func (h *HubSpotHandler) DisconnectAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	accountID := c.Param("id")

	var account models.HubSpotAccount
	if err := h.db.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "HubSpot account not found"})
		return
	}

	if err := h.db.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect HubSpot account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HubSpot account disconnected successfully"})
} 