package oauth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleOAuthProvider implements the OAuthProvider interface for Google authentication.
type GoogleOAuthProvider struct {
	config *oauth2.Config
}

// NewGoogleOAuthProvider creates a new instance of GoogleOAuthProvider with the given client ID, client secret, and redirect URL.
func NewGoogleOAuthProvider(clientID, clientSecret, redirectURL string) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		},
	}
}

// GetProviderName returns the name of the provider
func (g *GoogleOAuthProvider) GetProviderName() user.AuthProvider {
	return user.ProviderGoogle
}

// GetUserInfo exchanges the authorization code for an access token and retrieves the user's information from Google.
func (g *GoogleOAuthProvider) GetUserInfo(ctx context.Context, code string) (*service.OAuthUserInfo, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &service.OAuthUserInfo{
		Email:         userInfo.Email,
		UserName:      userInfo.Name,
		ProviderID:    userInfo.ID,
		EmailVerified: userInfo.VerifiedEmail,
	}, nil
}

// GetAuthURL generates the Google OAuth authorization URL with the given state parameter.
func (g *GoogleOAuthProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}
