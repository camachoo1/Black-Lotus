package oauth

import (
	"github.com/labstack/echo/v4"

	"black-lotus/internal/features/auth/oauth/github"
	"black-lotus/internal/features/auth/oauth/google"
)

type HandlerInterface interface {
	// GitHub OAuth endpoints
	GetGitHubAuthURL(ctx echo.Context) error
	HandleGitHubCallback(ctx echo.Context) error
	// Google OAuth endpoints
	GetGoogleAuthURL(ctx echo.Context) error
	HandleGoogleCallback(ctx echo.Context) error
}

// Handler manages OAuth authentication endpoints
type Handler struct {
	githubHandler github.HandlerInterface
	googleHandler google.HandlerInterface
}

// NewHandler creates a new OAuth handler
func NewHandler(
	githubHandler github.HandlerInterface,
	googleHandler google.HandlerInterface,
) *Handler {
	return &Handler{
		githubHandler: githubHandler,
		googleHandler: googleHandler,
	}
}

// GetGitHubAuthURL returns the GitHub OAuth URL
func (h *Handler) GetGitHubAuthURL(ctx echo.Context) error {
	return h.githubHandler.GetAuthURL(ctx)
}

// HandleGitHubCallback processes GitHub OAuth callback
func (h *Handler) HandleGitHubCallback(ctx echo.Context) error {
	return h.githubHandler.HandleCallback(ctx)
}

// GetGoogleAuthURL returns the Google OAuth URL
func (h *Handler) GetGoogleAuthURL(ctx echo.Context) error {
	return h.googleHandler.GetAuthURL(ctx)
}

// HandleGoogleCallback processes Google OAuth callback
func (h *Handler) HandleGoogleCallback(ctx echo.Context) error {
	return h.googleHandler.HandleCallback(ctx)
}
