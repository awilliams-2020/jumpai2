package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type AIService struct {
	client *openai.Client
}

func NewAIService() *AIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("Missing required OpenAI API key")
	}

	return &AIService{
		client: openai.NewClient(apiKey),
	}
}

func (s *AIService) EnrichAnswer(ctx context.Context, answer string, hubspotContact *HubSpotContact, linkedinProfile *LinkedInProfile) (string, error) {
	// Build context from HubSpot notes
	var hubspotContext strings.Builder
	if hubspotContact != nil {
		hubspotContext.WriteString("HubSpot Notes:\n")
		for _, note := range hubspotContact.Notes {
			hubspotContext.WriteString(fmt.Sprintf("- %s\n", note.Content))
		}
	}

	// Build context from LinkedIn profile
	var linkedinContext strings.Builder
	if linkedinProfile != nil {
		linkedinContext.WriteString("LinkedIn Profile:\n")
		linkedinContext.WriteString(fmt.Sprintf("Name: %s\n", linkedinProfile.Name))
		linkedinContext.WriteString(fmt.Sprintf("Title: %s\n", linkedinProfile.Title))
		linkedinContext.WriteString(fmt.Sprintf("Company: %s\n", linkedinProfile.Company))
		linkedinContext.WriteString(fmt.Sprintf("Location: %s\n", linkedinProfile.Location))
		linkedinContext.WriteString(fmt.Sprintf("About: %s\n", linkedinProfile.Description))
	}

	// Create the prompt for GPT
	prompt := fmt.Sprintf(`Given the following answer and context, provide an enriched version that includes relevant insights from the context:

Answer: %s

%s

%s

Please analyze the answer in the context of the provided information and add relevant insights or connections. Format the response as:

Original Answer: [the original answer]

Context: [relevant insights from HubSpot notes or LinkedIn profile]

Enriched Answer: [the answer with added context]`, 
		answer,
		hubspotContext.String(),
		linkedinContext.String())

	// Call OpenAI API
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an AI assistant that helps enrich answers with relevant context from HubSpot notes and LinkedIn profiles. Focus on finding meaningful connections and insights that add value to the original answer.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to get AI enrichment: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
} 