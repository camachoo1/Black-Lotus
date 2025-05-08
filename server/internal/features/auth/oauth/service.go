package oauth

import (
	"context"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/oauth/common"
	"black-lotus/internal/features/auth/oauth/github"
	"black-lotus/internal/features/auth/oauth/google"
)

// ServiceInterface defines the main OAuth service interface
type ServiceInterface interface {
	// GetAuthorizationURL returns the URL to redirect users for OAuth login
	GetAuthorizationURL(provider string, redirectURI string, state string) string

	// AuthenticateGitHub handles GitHub OAuth authentication
	AuthenticateGitHub(ctx context.Context, code string) (*models.User, error)

	// AuthenticateGoogle handles Google OAuth authentication
	AuthenticateGoogle(ctx context.Context, code string, redirectURI string) (*models.User, error)
}

// Service composes individual provider services into a unified interface
type Service struct {
	githubService *github.Service
	googleService *google.Service
}

// NewService creates a new OAuth service
func NewService(
	githubService *github.Service,
	googleService *google.Service,
) ServiceInterface {
	return &Service{
		githubService: githubService,
		googleService: googleService,
	}
}

// GetAuthorizationURL returns the URL to redirect users for OAuth login
func (s *Service) GetAuthorizationURL(provider string, redirectURI string, state string) string {
	return common.GetAuthorizationURL(provider, redirectURI, state)
}

// AuthenticateGitHub handles GitHub OAuth authentication
func (s *Service) AuthenticateGitHub(ctx context.Context, code string) (*models.User, error) {
	return s.githubService.Authenticate(ctx, code)
}

// AuthenticateGoogle handles Google OAuth authentication
func (s *Service) AuthenticateGoogle(ctx context.Context, code string, redirectURI string) (*models.User, error) {
	return s.googleService.Authenticate(ctx, code, redirectURI)
}
