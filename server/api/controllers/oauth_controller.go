// api/controllers/oauth_controller.go
package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"black-lotus/internal/services"
)

// OAuthController handles OAuth authentication endpoints
type OAuthController struct {
	oauthService   *services.OAuthService
	sessionService *services.SessionService
}

// NewOAuthController creates a new OAuth controller
func NewOAuthController(oauthService *services.OAuthService, sessionService *services.SessionService) *OAuthController {
	return &OAuthController{
		oauthService:   oauthService,
		sessionService: sessionService,
	}
}

// GetGitHubAuthURL returns the GitHub OAuth URL
func (c *OAuthController) GetGitHubAuthURL(ctx echo.Context) error {
	// Get base URL from request for redirect
	scheme := ctx.Scheme()
	host := ctx.Request().Host
	redirectURI := scheme + "://" + host + "/api/auth/github/callback"
	
	authURL := c.oauthService.GetAuthorizationURL("github", redirectURI)
	
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
	
	// Set cookie
	cookie := new(http.Cookie)
	cookie.Name = "session_token"
	cookie.Value = session.Token
	cookie.Expires = session.ExpiresAt
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	
	ctx.SetCookie(cookie)
	
	// Redirect to frontend
	return ctx.Redirect(http.StatusTemporaryRedirect, "/")
}

// GetGoogleAuthURL returns the Google OAuth URL
func (c *OAuthController) GetGoogleAuthURL(ctx echo.Context) error {
	// Get base URL from request for redirect
	scheme := ctx.Scheme()
	host := ctx.Request().Host
	redirectURI := scheme + "://" + host + "/api/auth/google/callback"
	
	authURL := c.oauthService.GetAuthorizationURL("google", redirectURI)
	
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
	
	// Get redirect URI (must match the one used to get auth URL)
	scheme := ctx.Scheme()
	host := ctx.Request().Host
	redirectURI := scheme + "://" + host + "/api/auth/google/callback"
	
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
	
	// Set cookie
	cookie := new(http.Cookie)
	cookie.Name = "session_token"
	cookie.Value = session.Token
	cookie.Expires = session.ExpiresAt
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	
	ctx.SetCookie(cookie)
	
	// Redirect to frontend
	return ctx.Redirect(http.StatusTemporaryRedirect, "/")
}