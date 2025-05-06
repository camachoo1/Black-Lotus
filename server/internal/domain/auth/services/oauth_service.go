package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"black-lotus/internal/domain/auth/repositories"
	"black-lotus/internal/models"
)

type OAuthServiceInterface interface {
	GetAuthorizationURL(provider string, redirectURI string, state string) string
	AuthenticateGitHub(ctx context.Context, code string) (*models.User, error)
	AuthenticateGoogle(ctx context.Context, code string, redirectURI string) (*models.User, error)
}

// OAuthService handles authentication with OAuth providers
type OAuthService struct {
	oauthRepo  repositories.OAuthRepositoryInterface
	userRepo   repositories.UserRepositoryInterface
	httpClient *http.Client
}

// GItHub OAuth responses
type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type githubUserResponse struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// Google OAuth responses
type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type googleUserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(
	oauthRepo repositories.OAuthRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
) *OAuthService {
	return &OAuthService{
		oauthRepo:  oauthRepo,
		userRepo:   userRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthorizationURL returns the URL to redirect users for OAuth login
func (s *OAuthService) GetAuthorizationURL(provider string, redirectURI string, state string) string {
	switch provider {
	case "github":
		return fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
			os.Getenv("GITHUB_CLIENT_ID"),
			url.QueryEscape(redirectURI),
			url.QueryEscape(state),
		)
	case "google":
		return fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s",
			os.Getenv("GOOGLE_CLIENT_ID"),
			url.QueryEscape(redirectURI),
			url.QueryEscape(state),
		)
	default:
		return ""
	}
}

// AuthenticateGitHub handles GitHub OAuth authentication
func (s *OAuthService) AuthenticateGitHub(ctx context.Context, code string) (*models.User, error) {
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

	var tokenResp githubTokenResponse
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

	var userResp githubUserResponse
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

		var emails []githubEmail
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

// AuthenticateGoogle handles Google OAuth authentication
func (s *OAuthService) AuthenticateGoogle(ctx context.Context, code string, redirectURI string) (*models.User, error) {
	// Exchange code for token
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))

	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp googleTokenResponse

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get user info
	userURL := "https://www.googleapis.com/oauth2/v1/userinfo"
	req, err = http.NewRequestWithContext(ctx, "GET", userURL, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	resp, err = s.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	defer resp.Body.Close()

	var userResp googleUserResponse

	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	// Check if OAuth account exists
	account, err := s.oauthRepo.GetOAuthAccount(ctx, "google", userResp.ID)

	if err == nil {
		// Account exists, get associated user
		user, err := s.userRepo.GetUserByID(ctx, account.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Update the token
		account.AccessToken = tokenResp.AccessToken
		account.RefreshToken = tokenResp.RefreshToken
		account.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
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
		ProviderID:     "google",
		ProviderUserID: userResp.ID,
		UserID:         user.ID,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	err = s.oauthRepo.CreateOAuthAccount(ctx, oauthAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth account: %w", err)
	}

	// Update user as verified if Google verified the email
	if !user.EmailVerified && userResp.VerifiedEmail {
		err = s.userRepo.SetEmailVerified(ctx, user.ID, true)
		if err != nil {
			// Non-critical error, log but continue
			fmt.Printf("failed to mark email as verified: %v", err)
		}
	}

	return user, nil
}
