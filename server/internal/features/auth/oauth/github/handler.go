package github

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"

	"black-lotus/internal/features/auth/session"
)

// Handler handles GitHub OAuth endpoints
type Handler struct {
	githubService  ServiceInterface
	sessionService session.ServiceInterface
}

type HandlerInterface interface {
	GetAuthURL(ctx echo.Context) error
	HandleCallback(ctx echo.Context) error
}

var _ HandlerInterface = (*Handler)(nil)

// NewHandler creates a new GitHub OAuth handler
func NewHandler(githubService ServiceInterface, sessionService session.ServiceInterface) *Handler {
	return &Handler{
		githubService:  githubService,
		sessionService: sessionService,
	}
}

// GetAuthURL returns the GitHub OAuth URL
func (h *Handler) GetAuthURL(ctx echo.Context) error {
	returnTo := ctx.QueryParam("returnTo")

	if returnTo == "" {
		returnTo = "/" // Default to home if not specified
	}

	// Get base URL from request for redirect
	scheme := ctx.Scheme()
	host := ctx.Request().Host

	// Important: Use redirect URI without query parameters
	redirectURI := fmt.Sprintf("%s://%s/api/auth/github/callback", scheme, host)

	// Pass returnTo as state parameter for security
	authURL := h.githubService.GetAuthURL(redirectURI, returnTo)

	return ctx.JSON(http.StatusOK, map[string]string{
		"url": authURL,
	})
}

// HandleCallback processes GitHub OAuth callback
func (h *Handler) HandleCallback(ctx echo.Context) error {
	// Get code from query parameters
	code := ctx.QueryParam("code")
	if code == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing code parameter",
		})
	}

	// Get state parameter (contains our returnTo value)
	state := ctx.QueryParam("state")
	returnTo := "/" // Default to callback page
	if state != "" {
		decodedState, err := url.QueryUnescape(state)
		if err == nil {
			// If we have a valid state, use it for returnTo
			returnTo = decodedState
		}
	}

	// Authenticate with GitHub
	user, err := h.githubService.Authenticate(ctx.Request().Context(), code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication failed: " + err.Error(),
		})
	}

	// Create session
	session, err := h.sessionService.CreateSession(ctx.Request().Context(), user.ID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create session",
		})
	}

	// Get frontend URL from environment or use default
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	redirectURL := frontendURL + "/auth/callback?returnTo=" + url.QueryEscape(returnTo)

	// Set access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = session.AccessToken
	accessCookie.Expires = session.AccessExpiry
	accessCookie.Path = "/"
	accessCookie.HttpOnly = true
	accessCookie.Secure = true
	accessCookie.SameSite = http.SameSiteLaxMode

	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = session.RefreshToken
	refreshCookie.Expires = session.RefreshExpiry
	refreshCookie.Path = "/"
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true
	refreshCookie.SameSite = http.SameSiteLaxMode

	ctx.SetCookie(accessCookie)
	ctx.SetCookie(refreshCookie)

	// Redirect to frontend
	return ctx.Redirect(http.StatusFound, redirectURL)
}
