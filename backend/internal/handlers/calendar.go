package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"github.com/yourusername/advisor-scheduling/internal/utils"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type CalendarHandler struct {
	db *gorm.DB
}

func NewCalendarHandler(db *gorm.DB) *CalendarHandler {
	return &CalendarHandler{
		db: db,
	}
}

// getConnectedCalendarEvents fetches events from all connected Google accounts for a user
func (h *CalendarHandler) getConnectedCalendarEvents(ctx context.Context, userID uint, startTime, endTime time.Time) ([]*calendar.Event, []string, error) {
	// Get user and their connected Google accounts
	var user models.User
	if err := h.db.Preload("GoogleAccounts").First(&user, userID).Error; err != nil {
		return nil, nil, fmt.Errorf("user not found: %v", err)
	}

	var allEvents []*calendar.Event
	var errors []string

	// Fetch events from all connected Google accounts
	if user.GoogleAccounts != nil && len(user.GoogleAccounts) > 0 {
		for _, account := range user.GoogleAccounts {
			if !account.IsActive {
				continue
			}

			// Create calendar service for this account
			client := utils.GetGoogleClient(ctx, account.AccessToken)
			srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to create calendar service for account %s: %v", account.Email, err))
				continue
			}

			// Get calendar IDs to fetch events from
			calendarIDs := account.CalendarIDs
			if len(calendarIDs) == 0 {
				calendarIDs = []string{"primary"} // Default to primary calendar if none specified
			}

			// Fetch events from all specified calendars for this account
			for _, calendarID := range calendarIDs {
				events, err := srv.Events.List(calendarID).
					TimeMin(startTime.Format(time.RFC3339)).
					TimeMax(endTime.Format(time.RFC3339)).
					SingleEvents(true).
					OrderBy("startTime").
					MaxResults(100). // Limit to 100 events per calendar
					Do()

				if err != nil {
					errors = append(errors, fmt.Sprintf("Failed to fetch events for calendar %s in account %s: %v", calendarID, account.Email, err))
					continue
				}

				if events.Items != nil {
					allEvents = append(allEvents, events.Items...)
				}
			}
		}
	}

	return allEvents, errors, nil
}

// GetCalendarEvents retrieves calendar events for the authenticated user
func (h *CalendarHandler) GetCalendarEvents(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Get time range from query parameters
	startTime := time.Now()
	endTime := startTime.AddDate(0, 1, 0) // Default to 1 month from now

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get user from database to retrieve AccessToken
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.AccessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token not found for user"})
		return
	}

	// Create calendar service for user's own account
	client := utils.GetGoogleClient(ctx, user.AccessToken)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create calendar service: %v", err)})
		return
	}

	// Fetch events from user's primary calendar
	userEvents, err := srv.Events.List("primary").
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(100).
		Do()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch user events: %v", err)})
		return
	}

	// Fetch events from connected accounts
	connectedEvents, errors, err := h.getConnectedCalendarEvents(ctx, userID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Combine user's events with connected account events
	var allEvents []*calendar.Event
	if userEvents.Items != nil {
		allEvents = append(allEvents, userEvents.Items...)
	}
	allEvents = append(allEvents, connectedEvents...)

	// Format events for response
	response := make([]gin.H, len(allEvents))
	for i, event := range allEvents {
		// Handle both date and dateTime fields
		var startTime, endTime string
		if event.Start.DateTime != "" {
			startTime = event.Start.DateTime
		} else {
			startTime = event.Start.Date + "T00:00:00Z"
		}
		if event.End.DateTime != "" {
			endTime = event.End.DateTime
		} else {
			endTime = event.End.Date + "T23:59:59Z"
		}

		response[i] = gin.H{
			"id":          event.Id,
			"summary":     event.Summary,
			"description": event.Description,
			"start_time":  startTime,
			"end_time":    endTime,
		}
	}

	// Add total count to response
	totalCount := len(allEvents)

	// Add any errors to the response
	if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"events": response,
			"errors": errors,
			"total":  totalCount,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": response,
		"total":  totalCount,
	})
} 