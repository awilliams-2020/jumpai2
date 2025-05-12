package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type HubSpotService struct {
	accessToken string
}

type HubSpotContact struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Company   string `json:"company"`
	Notes     []Note `json:"notes"`
}

type Note struct {
	Content string `json:"content"`
	Created string `json:"created"`
}

func NewHubSpotService() *HubSpotService {
	accessToken := os.Getenv("HUBSPOT_ACCESS_TOKEN")
	if accessToken == "" {
		panic("Missing required HubSpot access token")
	}

	return &HubSpotService{
		accessToken: accessToken,
	}
}

func (s *HubSpotService) FindContactByEmail(email string) (*HubSpotContact, error) {
	url := "https://api.hubapi.com/crm/v3/objects/contacts/search"
	
	searchBody := map[string]interface{}{
		"filterGroups": []map[string]interface{}{
			{
				"filters": []map[string]interface{}{
					{
						"propertyName": "email",
						"operator":     "EQ",
						"value":        email,
					},
				},
			},
		},
		"properties": []string{"email", "firstname", "lastname", "company"},
	}

	jsonBody, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hubspot API error: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			ID         string `json:"id"`
			Properties struct {
				Email     string `json:"email"`
				FirstName string `json:"firstname"`
				LastName  string `json:"lastname"`
				Company   string `json:"company"`
			} `json:"properties"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(result.Results) == 0 {
		return nil, nil
	}

	contact := result.Results[0]
	
	// Get notes for the contact
	notes, err := s.getContactNotes(contact.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact notes: %v", err)
	}

	return &HubSpotContact{
		ID:        contact.ID,
		Email:     contact.Properties.Email,
		FirstName: contact.Properties.FirstName,
		LastName:  contact.Properties.LastName,
		Company:   contact.Properties.Company,
		Notes:     notes,
	}, nil
}

func (s *HubSpotService) getContactNotes(contactID string) ([]Note, error) {
	url := fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/notes/search")
	
	searchBody := map[string]interface{}{
		"filterGroups": []map[string]interface{}{
			{
				"filters": []map[string]interface{}{
					{
						"propertyName": "hs_timestamp",
						"operator":     "GTE",
						"value":        "0",
					},
				},
			},
		},
		"properties": []string{"hs_note_body", "hs_timestamp"},
		"associations": []map[string]interface{}{
			{
				"to": map[string]interface{}{
					"id": contactID,
				},
				"types": []map[string]interface{}{
					{
						"associationCategory": "HUBSPOT_DEFINED",
						"associationTypeId":   1,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notes request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create notes request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send notes request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hubspot notes API error: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Properties struct {
				Content string `json:"hs_note_body"`
				Created string `json:"hs_timestamp"`
			} `json:"properties"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode notes response: %v", err)
	}

	notes := make([]Note, len(result.Results))
	for i, r := range result.Results {
		notes[i] = Note{
			Content: r.Properties.Content,
			Created: r.Properties.Created,
		}
	}

	return notes, nil
} 