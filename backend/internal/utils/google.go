package utils

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// GetGoogleClient creates a new HTTP client with the provided access token
func GetGoogleClient(ctx context.Context, accessToken string) *http.Client {
	token := &oauth2.Token{
		AccessToken: accessToken,
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
} 