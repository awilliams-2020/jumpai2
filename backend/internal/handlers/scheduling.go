package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"github.com/yourusername/advisor-scheduling/internal/services"
	"gorm.io/gorm"
)

type SchedulingHandler struct {
	db *gorm.DB
	emailService *services.EmailService
}

func NewSchedulingHandler(db *gorm.DB, emailService *services.EmailService) *SchedulingHandler {
	return &SchedulingHandler{
		db: db,
		emailService: emailService,
	}
}

// CreateSchedulingLink creates a new scheduling link
func (h *SchedulingHandler) CreateSchedulingLink(c *gin.Context) {
	var input struct {
		Title            string     `json:"title" binding:"required"`
		Duration         int        `json:"duration" binding:"required"`
		MaxUses          *int       `json:"max_uses"`
		ExpiresAt        *time.Time `json:"expires_at"`
		MaxDaysInAdvance int        `json:"max_days_in_advance" binding:"required"`
		CustomQuestions  []string   `json:"custom_questions" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert custom questions to JSON string
	jsonBytes, err := json.Marshal(input.CustomQuestions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process custom questions"})
		return
	}
	customQuestionsJSON := string(jsonBytes)

	userID := c.GetUint("user_id")
	link := &models.SchedulingLink{
		UserID:           userID,
		Title:            input.Title,
		Duration:         input.Duration,
		MaxUses:          input.MaxUses,
		ExpiresAt:        input.ExpiresAt,
		MaxDaysInAdvance: input.MaxDaysInAdvance,
		CustomQuestions:  customQuestionsJSON,
		IsActive:         true,
	}

	if err := h.db.Create(link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduling link"})
		return
	}

	// Parse the custom questions back to array for the response
	var customQuestions []string
	if err := json.Unmarshal([]byte(link.CustomQuestions), &customQuestions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process custom questions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                  link.ID,
		"title":              link.Title,
		"duration":           link.Duration,
		"max_uses":           link.MaxUses,
		"expires_at":         link.ExpiresAt,
		"max_days_in_advance": link.MaxDaysInAdvance,
		"custom_questions":   customQuestions,
		"is_active":          link.IsActive,
	})
}

// GetSchedulingLink retrieves a scheduling link by ID
func (h *SchedulingHandler) GetSchedulingLink(c *gin.Context) {
	id := c.Param("id")
	var link models.SchedulingLink

	if err := h.db.First(&link, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduling link not found"})
		return
	}

	// Parse custom questions from JSON string
	var customQuestions []string
	if link.CustomQuestions != "" {
		if err := json.Unmarshal([]byte(link.CustomQuestions), &customQuestions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process custom questions"})
			return
		}
	}

	// Return response in snake_case format
	c.JSON(http.StatusOK, gin.H{
		"id":                  link.ID,
		"title":              link.Title,
		"duration":           link.Duration,
		"max_uses":           link.MaxUses,
		"expires_at":         link.ExpiresAt,
		"max_days_in_advance": link.MaxDaysInAdvance,
		"custom_questions":   customQuestions,
		"is_active":          link.IsActive,
	})
}

// CreateSchedulingWindow creates a new scheduling window
func (h *SchedulingHandler) CreateSchedulingWindow(c *gin.Context) {
	var input struct {
		StartHour int `json:"start_hour" binding:"required"`
		EndHour   int `json:"end_hour" binding:"required"`
		Weekday   int `json:"weekday" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	window := &models.SchedulingWindow{
		UserID:    userID,
		StartHour: input.StartHour,
		EndHour:   input.EndHour,
		Weekday:   input.Weekday,
		IsActive:  true,
	}

	if err := h.db.Create(window).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduling window"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         window.ID,
		"start_hour": window.StartHour,
		"end_hour":   window.EndHour,
		"weekday":    window.Weekday,
		"is_active":  window.IsActive,
	})
}

// GetAvailableSlots retrieves available time slots for a scheduling link
func (h *SchedulingHandler) GetAvailableSlots(c *gin.Context) {
	linkID := c.Param("id")
	var link models.SchedulingLink

	if err := h.db.First(&link, linkID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduling link not found"})
		return
	}

	// Parse date from query param
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing date parameter (expected format: yyyy-mm-dd)"})
		return
	}
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (expected: yyyy-mm-dd)"})
		return
	}

	// Check expiration
	if link.ExpiresAt != nil && selectedDate.After(*link.ExpiresAt) {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	// Check max days in advance
	today := time.Now().Truncate(24 * time.Hour)
	maxDate := today.AddDate(0, 0, link.MaxDaysInAdvance)
	if selectedDate.After(maxDate) {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	// Get user's scheduling windows for the selected weekday
	var windows []models.SchedulingWindow
	if err := h.db.Where("user_id = ? AND is_active = ? AND weekday = ?", link.UserID, true, int(selectedDate.Weekday())).Find(&windows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scheduling windows"})
		return
	}
	if len(windows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active scheduling windows found for this day"})
		return
	}

	// Get existing meetings for this link on the selected date
	var meetings []models.Meeting
	startOfDay := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), 0, 0, 0, 0, selectedDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	if err := h.db.Where("scheduling_link_id = ? AND start_time >= ? AND start_time < ?", link.ID, startOfDay, endOfDay).Find(&meetings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meetings"})
		return
	}

	// Build a list of all possible slots for the day
	slots := []gin.H{}
	meetingDuration := time.Duration(link.Duration) * time.Minute
	for _, window := range windows {
		windowStart := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), window.StartHour, 0, 0, 0, selectedDate.Location())
		windowEnd := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), window.EndHour, 0, 0, 0, selectedDate.Location())
		for slotStart := windowStart; slotStart.Add(meetingDuration).UTC().Before(windowEnd.UTC()) || slotStart.Add(meetingDuration).UTC().Equal(windowEnd.UTC()); slotStart = slotStart.Add(meetingDuration) {
			slotEnd := slotStart.Add(meetingDuration)
			// Check for overlap with existing meetings
			overlaps := false
			for _, meeting := range meetings {
				if (slotStart.Before(meeting.EndTime) && slotEnd.After(meeting.StartTime)) || slotStart.Equal(meeting.StartTime) {
					overlaps = true
					break
				}
			}
			if !overlaps && slotStart.After(time.Now()) {
				slots = append(slots, gin.H{
					"start": slotStart.UTC().Format(time.RFC3339),
					"end": slotEnd.UTC().Format(time.RFC3339),
				})
			}
		}
	}

	// If max_uses is set, check if the number of meetings already scheduled meets the limit
	if link.MaxUses != nil {
		var totalMeetings int64
		h.db.Model(&models.Meeting{}).Where("scheduling_link_id = ?", link.ID).Count(&totalMeetings)
		if int(totalMeetings) >= *link.MaxUses {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
	}

	c.JSON(http.StatusOK, slots)
}

// GetSchedulingLinks retrieves all scheduling links for the authenticated user
func (h *SchedulingHandler) GetSchedulingLinks(c *gin.Context) {
	userID := c.GetUint("user_id")
	var links []models.SchedulingLink

	if err := h.db.Where("user_id = ?", userID).Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scheduling links"})
		return
	}

	// Convert links to response format with proper JSON handling
	response := make([]gin.H, len(links))
	for i, link := range links {
		var customQuestions []string
		if link.CustomQuestions != "" {
			if err := json.Unmarshal([]byte(link.CustomQuestions), &customQuestions); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process custom questions"})
				return
			}
		}

		response[i] = gin.H{
			"id":                  link.ID,
			"title":              link.Title,
			"duration":           link.Duration,
			"max_uses":           link.MaxUses,
			"expires_at":         link.ExpiresAt,
			"max_days_in_advance": link.MaxDaysInAdvance,
			"custom_questions":   customQuestions,
			"is_active":          link.IsActive,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetSchedulingWindows retrieves all scheduling windows for the authenticated user
func (h *SchedulingHandler) GetSchedulingWindows(c *gin.Context) {
	userID := c.GetUint("user_id")
	var windows []models.SchedulingWindow

	if err := h.db.Where("user_id = ?", userID).Find(&windows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scheduling windows"})
		return
	}

	// Convert windows to response format with snake_case
	response := make([]gin.H, len(windows))
	for i, window := range windows {
		response[i] = gin.H{
			"id":         window.ID,
			"start_hour": window.StartHour,
			"end_hour":   window.EndHour,
			"weekday":    window.Weekday,
			"is_active":  window.IsActive,
		}
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSchedulingWindow deletes a scheduling window
func (h *SchedulingHandler) DeleteSchedulingWindow(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	// First check if the window exists and belongs to the user
	var window models.SchedulingWindow
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&window).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduling window not found"})
		return
	}

	// Soft delete by setting is_active to false
	if err := h.db.Model(&window).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheduling window"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduling window deleted successfully"})
}

// CreateMeeting creates a new meeting and updates the scheduling link's max uses
func (h *SchedulingHandler) CreateMeeting(c *gin.Context) {
	linkID := c.Param("id")
	var link models.SchedulingLink

	// Get the scheduling link
	if err := h.db.First(&link, linkID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduling link not found"})
		return
	}

	// Check if link is still active
	if !link.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This scheduling link is no longer active"})
		return
	}

	// Check if link has expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This scheduling link has expired"})
		return
	}

	// Check max uses
	if link.MaxUses != nil {
		var totalMeetings int64
		h.db.Model(&models.Meeting{}).Where("scheduling_link_id = ?", link.ID).Count(&totalMeetings)
		if int(totalMeetings) >= *link.MaxUses {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This scheduling link has reached its maximum number of uses"})
			return
		}
	}

	var input struct {
		ClientEmail  string            `json:"client_email" binding:"required,email"`
		LinkedInURL  string            `json:"linkedin_url"`
		StartTime    time.Time         `json:"start_time" binding:"required"`
		EndTime      time.Time         `json:"end_time" binding:"required"`
		Answers      map[string]string `json:"answers" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate time slot
	if input.StartTime.After(input.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time slot"})
		return
	}

	// Check if the time slot is still available
	var existingMeeting models.Meeting
	result := h.db.Where(
		"scheduling_link_id = ? AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		link.ID, input.StartTime, input.StartTime, input.EndTime, input.EndTime,
	).First(&existingMeeting)

	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			// Only log actual errors, not the "no record found" case
			c.Error(fmt.Errorf("error checking time slot availability: %v", result.Error))
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This time slot is no longer available"})
		return
	}

	// Convert answers to string array
	answers := make(models.StringSlice, 0, len(input.Answers))
	for question, answer := range input.Answers {
		answers = append(answers, question+": "+answer)
	}

	// Create the meeting
	meeting := &models.Meeting{
		SchedulingLinkID: link.ID,
		UserID:          link.UserID,
		ClientEmail:     input.ClientEmail,
		LinkedInURL:     input.LinkedInURL,
		StartTime:       input.StartTime,
		EndTime:         input.EndTime,
		Answers:         answers,
		LinkedInData:    "{}", // Initialize with empty JSON object
	}

	if err := h.db.Create(meeting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meeting"})
		return
	}

	// Find and deactivate the scheduling window that contains this meeting
	var window models.SchedulingWindow
	meetingWeekday := int(input.StartTime.Weekday())
	meetingHour := input.StartTime.Hour()
	if err := h.db.Where(
		"user_id = ? AND weekday = ? AND start_hour <= ? AND end_hour > ? AND is_active = ?",
		link.UserID, meetingWeekday, meetingHour, meetingHour, true,
	).First(&window).Error; err == nil {
		// Update the window's is_active status
		if err := h.db.Model(&window).Update("is_active", false).Error; err != nil {
			// Log the error but don't fail the meeting creation
			c.Error(fmt.Errorf("failed to update scheduling window: %v", err))
		}
	}

	// Get the user's email to send notification
	var user models.User
	if err := h.db.First(&user, link.UserID).Error; err != nil {
		// Log the error but don't fail the meeting creation
		c.Error(fmt.Errorf("failed to fetch user for email notification: %v", err))
	} else {
		// Send email notification in a goroutine
		go func() {
			// Create a new context with timeout for the entire email process
			emailCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			meetingDetails := map[string]interface{}{
				"meeting_id":   meeting.ID,
				"client_email": input.ClientEmail,
				"linkedin_url": input.LinkedInURL,
				"start_time":   input.StartTime.Format(time.RFC3339),
				"end_time":     input.EndTime.Format(time.RFC3339),
				"answers":      answers,
			}
			if err := h.emailService.SendMeetingNotification(emailCtx, user.Email, meetingDetails); err != nil {
				// Log the error but don't fail the meeting creation
				fmt.Printf("Failed to send email notification: %v\n", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            meeting.ID,
		"client_email":  meeting.ClientEmail,
		"linkedin_url":  meeting.LinkedInURL,
		"start_time":    meeting.StartTime,
		"end_time":      meeting.EndTime,
		"answers":       input.Answers,
	})
}

// GetLinkMeetings retrieves all meetings for a specific scheduling link
func (h *SchedulingHandler) GetLinkMeetings(c *gin.Context) {
	linkID := c.Param("id")
	var meetings []models.Meeting

	if err := h.db.Where("scheduling_link_id = ?", linkID).Find(&meetings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meetings"})
		return
	}

	// Convert meetings to response format
	response := make([]gin.H, len(meetings))
	for i, meeting := range meetings {
		response[i] = gin.H{
			"id":            meeting.ID,
			"client_email":  meeting.ClientEmail,
			"linkedin_url":  meeting.LinkedInURL,
			"start_time":    meeting.StartTime,
			"end_time":      meeting.EndTime,
			"answers":       meeting.Answers,
			"context_notes": meeting.ContextNotes,
		}
	}

	c.JSON(http.StatusOK, response)
}
