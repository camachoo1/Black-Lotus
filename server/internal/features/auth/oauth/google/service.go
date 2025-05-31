package google

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

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/oauth/common"
)

type ServiceInterface interface {
	Authenticate(ctx context.Context, code string, redirectURI string) (*models.User, error)
	GetAuthURL(redirectURI string, state string) string
}

// Service handles Google OAuth authentication
type Service struct {
	oauthRepo  OAuthRepository
	userRepo   UserRepository
	httpClient *http.Client
}

// NewService creates a new Google OAuth service
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

func (s *Service) GetAuthURL(redirectUri string, state string) string {
	return common.GetAuthorizationURL("google", redirectUri, state)
}

// Authenticate handles Google OAuth authentication
func (s *Service) Authenticate(ctx context.Context, code string, redirectURI string) (*models.User, error) {
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
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

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

	var userResp struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

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
