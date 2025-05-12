package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/yourusername/advisor-scheduling/internal/models"
	"gorm.io/gorm"
)

type EmailService struct {
	client      *sendgrid.Client
	from        *mail.Email
	hubspot     *HubSpotService
	linkedin    *LinkedInService
	ai          *AIService
	db          *gorm.DB
}

func NewEmailService(db *gorm.DB) *EmailService {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
	fromName := os.Getenv("SENDGRID_FROM_NAME")

	if apiKey == "" || fromEmail == "" || fromName == "" {
		panic("Missing required SendGrid environment variables")
	}

	return &EmailService{
		client:   sendgrid.NewSendClient(apiKey),
		from:     mail.NewEmail(fromName, fromEmail),
		hubspot:  NewHubSpotService(),
		linkedin: NewLinkedInService(),
		ai:       NewAIService(),
		db:       db,
	}
}

func (s *EmailService) SendMeetingNotification(ctx context.Context, toEmail string, meetingDetails map[string]interface{}) error {
	to := mail.NewEmail("", toEmail)
	subject := "New Meeting Scheduled"

	// Try to find the contact in HubSpot first
	contact, err := s.hubspot.FindContactByEmail(meetingDetails["client_email"].(string))
	if err != nil {
		fmt.Printf("Failed to find HubSpot contact: %v\n", err)
	}

	// Only scrape LinkedIn if we don't have enough context from HubSpot
	var linkedinProfile *LinkedInProfile
	hasEnoughContext := false
	if contact != nil && len(contact.Notes) > 0 {
		// Check if we have any recent or relevant notes
		for _, note := range contact.Notes {
			if strings.Contains(strings.ToLower(note.Content), "concern") ||
				strings.Contains(strings.ToLower(note.Content), "goal") ||
				strings.Contains(strings.ToLower(note.Content), "need") {
				hasEnoughContext = true
				break
			}
		}
	}

	// If we don't have enough context from HubSpot, try LinkedIn
	if !hasEnoughContext {
		if linkedinURL, ok := meetingDetails["linkedin_url"].(string); ok && linkedinURL != "" {
			// Create a new context with timeout for LinkedIn scraping
			linkedinCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()
			
			profile, err := s.linkedin.ScrapeProfile(linkedinCtx, linkedinURL)
			if err != nil {
				if err == context.DeadlineExceeded {
					fmt.Printf("LinkedIn scraping timed out after 2 minutes for URL: %s\n", linkedinURL)
				} else {
					fmt.Printf("Failed to scrape LinkedIn profile for URL %s: %v\n", linkedinURL, err)
				}
				// Return a basic profile with the URL
				linkedinProfile = &LinkedInProfile{
					Name:        "LinkedIn Profile",
					Title:       "Profile URL: " + linkedinURL,
					Description: "Profile scraping timed out or failed. Please check the URL manually.",
				}
			} else {
				linkedinProfile = profile
			}
		}
	}

	// Format the meeting details into a readable message
	content := fmt.Sprintf(`
New Meeting Scheduled

Client Email: %s
LinkedIn URL: %s
Start Time: %s
End Time: %s

Questions and Answers:
`, 
		meetingDetails["client_email"],
		meetingDetails["linkedin_url"],
		meetingDetails["start_time"],
		meetingDetails["end_time"])

	// Process and enrich answers
	answers := meetingDetails["answers"].(models.StringSlice)
	enrichedAnswers := make(map[string]string)
	
	for _, answer := range answers {
		// Split the answer into question and answer parts
		parts := strings.SplitN(answer, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		question := parts[0]
		answerText := parts[1]

		// Create a new context with timeout for AI enrichment
		aiCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Enrich the answer with AI
		enrichedAnswer, err := s.ai.EnrichAnswer(aiCtx, answerText, contact, linkedinProfile)
		if err != nil {
			fmt.Printf("Failed to enrich answer: %v\n", err)
			enrichedAnswer = answerText
		}
		enrichedAnswers[question] = enrichedAnswer
		content += fmt.Sprintf("\n%s\nContext: %s\n", question, enrichedAnswer)
	}

	// Store enriched answers in the meeting's context_notes
	if meetingID, ok := meetingDetails["meeting_id"].(uint); ok {
		contextNotes, err := json.Marshal(enrichedAnswers)
		if err != nil {
			fmt.Printf("Failed to marshal enriched answers: %v\n", err)
		} else {
			if err := s.db.Model(&models.Meeting{}).Where("id = ?", meetingID).Update("context_notes", string(contextNotes)).Error; err != nil {
				fmt.Printf("Failed to update meeting context notes: %v\n", err)
			}
		}
	}

	// Create the email message
	message := mail.NewSingleEmail(s.from, subject, to, content, content)

	// Send the email
	response, err := s.client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid API error: %s", response.Body)
	}

	return nil
} 