package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar.events.readonly",
		},
		Endpoint: google.Endpoint,
	}

	googleConnectConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CONNECT_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar.events.readonly",
		},
		Endpoint: google.Endpoint,
	}
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// GoogleLogin initiates the OAuth 2.0 flow
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := googleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the OAuth 2.0 callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	token, err := googleOAuthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info from Google
	client := googleOAuthConfig.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Find or create user
	var user models.User
	result := h.db.Where("email = ?", userInfo.Email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			user = models.User{
				Email:         userInfo.Email,
				Name:          userInfo.Name,
				ProfilePicture: userInfo.Picture,
				GoogleID:      userInfo.ID,
				HubspotID:     nil,
				IsActive:      true,
				LastLoginAt:   time.Now(),
			}
			if err := h.db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// Update user's tokens and info
	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken
	user.TokenExpiry = token.Expiry
	user.LastLoginAt = time.Now()
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Generate JWT token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Redirect to frontend with token
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Default Vite dev server URL
	}
	// Redirect back to dashboard
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/login?token="+tokenString)
}

// ConnectGoogleAccount initiates the OAuth 2.0 flow for connecting an additional Google account
func (h *GoogleHandler) ConnectGoogleAccount(c *gin.Context) {
	// Get token from query parameter
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token required"})
		return
	}

	// Store the token in a cookie for the callback
	c.SetCookie("auth_token", tokenString, 300, "/", "", false, true)

	// Redirect to Google OAuth
	url := googleConnectConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// ConnectGoogleAccountCallback handles the OAuth 2.0 callback for connecting an additional Google account
func (h *GoogleHandler) ConnectGoogleAccountCallback(c *gin.Context) {
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

	code := c.Query("code")
	googleToken, err := googleConnectConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info from Google
	client := googleConnectConfig.Client(c, googleToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Check if this Google account is already connected
	var existingAccount models.GoogleAccount
	result := h.db.Where("google_id = ?", userInfo.ID).First(&existingAccount)
	if result.Error == nil {
		// Account already exists, update it
		existingAccount.AccessToken = googleToken.AccessToken
		existingAccount.RefreshToken = googleToken.RefreshToken
		existingAccount.TokenExpiry = googleToken.Expiry
		existingAccount.LastSyncAt = time.Now()
		existingAccount.IsActive = true
		existingAccount.ProfilePicture = userInfo.Picture
		existingAccount.Name = userInfo.Name
		if err := h.db.Save(&existingAccount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing account"})
			return
		}
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Create new Google account
		newAccount := models.GoogleAccount{
			UserID:         uint(userID),
			GoogleID:       userInfo.ID,
			Email:          userInfo.Email,
			AccessToken:    googleToken.AccessToken,
			RefreshToken:   googleToken.RefreshToken,
			TokenExpiry:    googleToken.Expiry,
			CalendarIDs:    models.StringSlice{"primary"}, // Default to primary calendar
			IsActive:       true,
			LastSyncAt:     time.Now(),
			ProfilePicture: userInfo.Picture,
			Name:           userInfo.Name,
		}

		if err := h.db.Create(&newAccount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Google account"})
			return
		}
	}

	// Clear the auth token cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	// Redirect back to dashboard
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Default Vite dev server URL
	}
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/dashboard")
}

// Profile returns the user's profile information
func (h *AuthHandler) Profile(c *gin.Context) {
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