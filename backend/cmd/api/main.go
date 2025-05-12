package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourusername/advisor-scheduling/internal/handlers"
	"github.com/yourusername/advisor-scheduling/internal/middleware"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"github.com/yourusername/advisor-scheduling/internal/services"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.SchedulingWindow{},
		&models.SchedulingLink{},
		&models.Meeting{},
		&models.GoogleAccount{},
		&models.HubSpotAccount{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	emailService := services.NewEmailService(db)
	schedulingHandler := handlers.NewSchedulingHandler(db, emailService)
	hubspotHandler := handlers.NewHubSpotHandler(db)
	googleHandler := handlers.NewGoogleHandler(db)
	calendarHandler := handlers.NewCalendarHandler(db)
	// Setup router
	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORS())

	// Public routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/scheduling/links/:id/public", schedulingHandler.GetPublicSchedulingLink)
	router.GET("/scheduling/links/:id/slots/public", schedulingHandler.GetPublicAvailableSlots)
	router.POST("/scheduling/links/:id/meetings/public", schedulingHandler.CreatePublicMeeting)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.Auth(db))
	{
		// Scheduling routes
		scheduling := protected.Group("/scheduling")
		{
			scheduling.POST("/links", schedulingHandler.CreateSchedulingLink)
			scheduling.GET("/links", schedulingHandler.GetSchedulingLinks)
			scheduling.GET("/links/:id", schedulingHandler.GetSchedulingLink)
			scheduling.GET("/links/:id/slots", schedulingHandler.GetAvailableSlots)
			scheduling.GET("/links/:id/meetings", schedulingHandler.GetLinkMeetings)
			scheduling.POST("/windows", schedulingHandler.CreateSchedulingWindow)
			scheduling.GET("/windows", schedulingHandler.GetSchedulingWindows)
			scheduling.DELETE("/windows/:id", schedulingHandler.DeleteSchedulingWindow)
		}

		// Google routes
		google := protected.Group("/google")
		{
			google.GET("/profile", googleHandler.Profile)
			google.GET("/connect", googleHandler.ConnectGoogleAccount)
			google.GET("/accounts", googleHandler.GetGoogleAccounts)
			google.DELETE("/accounts/:id", googleHandler.DisconnectGoogleAccount)
			google.GET("/calendar/events", calendarHandler.GetCalendarEvents)
		}

		// HubSpot routes
		hubspot := protected.Group("/hubspot")
		{
			hubspot.GET("/connect", hubspotHandler.HubSpotConnect)
			hubspot.GET("/accounts", hubspotHandler.GetHubSpotAccounts)
			hubspot.DELETE("/accounts/:id", hubspotHandler.DisconnectAccount)
		}
	}

	// Auth routes
	router.GET("/auth/google/login", authHandler.GoogleLogin)
	router.GET("/auth/google/callback", authHandler.GoogleCallback)
	router.GET("/auth/google/connect/callback", googleHandler.ConnectGoogleAccountCallback)
	router.GET("/auth/hubspot/connect/callback", hubspotHandler.HubSpotConnectCallback)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
