package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"

	"black-lotus/internal/domain/auth/services"
)

type OAuthControllerInterface interface {
	GetGitHubAuthURL(ctx echo.Context) error
	HandleGitHubCallback(ctx echo.Context) error
	GetGoogleAuthURL(ctx echo.Context) error
	HandleGoogleCallback(ctx echo.Context) error
}

// OAuthController handles OAuth authentication endpoints
type OAuthController struct {
	oauthService   services.OAuthServiceInterface
	sessionService services.SessionServiceInterface
}

// NewOAuthController creates a new OAuth controller
func NewOAuthController(oauthService services.OAuthServiceInterface, sessionService services.SessionServiceInterface) *OAuthController {
	return &OAuthController{
		oauthService:   oauthService,
		sessionService: sessionService,
	}
}

// GetGitHubAuthURL returns the GitHub OAuth URL
func (c *OAuthController) GetGitHubAuthURL(ctx echo.Context) error {
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
	authURL := c.oauthService.GetAuthorizationURL("github", redirectURI, returnTo)

	return ctx.JSON(http.StatusOK, map[string]string{
		"url": authURL,
	})
}

// HandleGitHubCallback processes GitHub OAuth callback
func (c *OAuthController) HandleGitHubCallback(ctx echo.Context) error {
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
	user, err := c.oauthService.AuthenticateGitHub(ctx.Request().Context(), code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication failed: " + err.Error(),
		})
	}

	// Create session
	session, err := c.sessionService.CreateSession(ctx.Request().Context(), user.ID)
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

// GetGoogleAuthURL returns the Google OAuth URL
func (c *OAuthController) GetGoogleAuthURL(ctx echo.Context) error {
	returnTo := ctx.QueryParam("returnTo")

	if returnTo == "" {
		returnTo = "/" // Default to home if not specified
	}

	// Get base URL from request for redirect
	scheme := ctx.Scheme()
	host := ctx.Request().Host

	// Important: Use redirect URI without query parameters
	redirectURI := fmt.Sprintf("%s://%s/api/auth/google/callback", scheme, host)

	// Pass returnTo as state parameter for security
	authURL := c.oauthService.GetAuthorizationURL("google", redirectURI, returnTo)

	return ctx.JSON(http.StatusOK, map[string]string{
		"url": authURL,
	})
}

// HandleGoogleCallback processes Google OAuth callback
func (c *OAuthController) HandleGoogleCallback(ctx echo.Context) error {
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

	// Get redirect URI (must match the one used to get auth URL)
	scheme := ctx.Scheme()
	host := ctx.Request().Host
	redirectURI := fmt.Sprintf("%s://%s/api/auth/google/callback", scheme, host)

	// Authenticate with Google
	user, err := c.oauthService.AuthenticateGoogle(ctx.Request().Context(), code, redirectURI)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication failed: " + err.Error(),
		})
	}

	// Create session
	session, err := c.sessionService.CreateSession(ctx.Request().Context(), user.ID)
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
	accessCookie.SameSite = http.SameSiteLaxMode // Critical for OAuth

	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = session.RefreshToken
	refreshCookie.Expires = session.RefreshExpiry
	refreshCookie.Path = "/"
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true
	refreshCookie.SameSite = http.SameSiteLaxMode // Critical for OAuth

	ctx.SetCookie(accessCookie)
	ctx.SetCookie(refreshCookie)

	// Redirect to frontend
	return ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
