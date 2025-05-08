package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/oauth/common"
)

type ServiceInterface interface {
	Authenticate(ctx context.Context, code string) (*models.User, error)
	GetAuthURL(redirectURI string, state string) string
}

// Service handles GitHub OAuth authentication
type Service struct {
	oauthRepo  OAuthRepository
	userRepo   UserRepository
	httpClient *http.Client
}

// NewService creates a new GitHub OAuth service
func NewService(
	oauthRepo OAuthRepository,
	userRepo UserRepository,
) *Service {
	return &Service{
		oauthRepo:  oauthRepo,
		userRepo:   userRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) GetAuthURL(redirectURI string, state string) string {
	return common.GetAuthorizationURL("github", redirectURI, state)
}

// Authenticate handles GitHub OAuth authentication
func (s *Service) Authenticate(ctx context.Context, code string) (*models.User, error) {
	// Exchange code for token
	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get user info
	userURL := "https://api.github.com/user"
	req, err = http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", "token "+tokenResp.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userResp struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	// Get user email if not provided
	if userResp.Email == "" {
		emailURL := "https://api.github.com/user/emails"
		req, err = http.NewRequestWithContext(ctx, "GET", emailURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create email request: %w", err)
		}
		req.Header.Set("Authorization", "token "+tokenResp.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err = s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get user emails: %w", err)
		}
		defer resp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			return nil, fmt.Errorf("failed to parse emails response: %w", err)
		}

		// Find primary verified email
		for _, email := range emails {
			if email.Primary && email.Verified {
				userResp.Email = email.Email
				break
			}
		}

		if userResp.Email == "" && len(emails) > 0 {
			userResp.Email = emails[0].Email
		}
	}

	if userResp.Email == "" {
		return nil, fmt.Errorf("no email provided by GitHub")
	}

	// Check if OAuth account exists
	providerUserID := fmt.Sprintf("%d", userResp.ID)
	account, err := s.oauthRepo.GetOAuthAccount(ctx, "github", providerUserID)

	if err == nil {
		// Account exists, get associated user
		user, err := s.userRepo.GetUserByID(ctx, account.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Update the token
		account.AccessToken = tokenResp.AccessToken
		err = s.oauthRepo.CreateOAuthAccount(ctx, *account)
		if err != nil {
			return nil, fmt.Errorf("failed to update OAuth account: %w", err)
		}

		return user, nil
	}

	// Check if user with this email exists
	user, err := s.userRepo.GetUserByEmail(ctx, userResp.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}

	// If user is nil, create a new one
	if user == nil {
		input := models.CreateUserInput{
			Name:  userResp.Name,
			Email: userResp.Email,
		}

		user, err = s.userRepo.CreateUser(ctx, input, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Create OAuth account
	oauthAccount := models.OAuthAccount{
		ProviderID:     "github",
		ProviderUserID: providerUserID,
		UserID:         user.ID,
		AccessToken:    tokenResp.AccessToken,
		ExpiresAt:      time.Now().Add(24 * time.Hour), // GitHub tokens don't expire by default
	}

	err = s.oauthRepo.CreateOAuthAccount(ctx, oauthAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth account: %w", err)
	}

	// Update user as verified
	if !user.EmailVerified {
		err = s.userRepo.SetEmailVerified(ctx, user.ID, true)
		if err != nil {
			// Non-critical error, log but continue
			fmt.Printf("failed to mark email as verified: %v", err)
		}
	}

	return user, nil
}
