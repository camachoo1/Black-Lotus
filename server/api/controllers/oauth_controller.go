package controllers

import (
	"fmt"
	"net/http"
	"net/url"

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
	returnTo := ctx.QueryParam("returnTo")

	if returnTo == "" {
			returnTo = "/" // Default to home if not specified
	}
	// Get base URL from request for redirect
	scheme := ctx.Scheme()
	host := ctx.Request().Host
	redirectURI := fmt.Sprintf("%s://%s/api/auth/github/callback?returnTo=%s", 
    scheme, host, url.QueryEscape(returnTo))
	
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

	// Get returnTo from query params
	returnTo := ctx.QueryParam("returnTo")
	if returnTo == "" {
			returnTo = "/" // Default to home
	}

	redirectURL := "http://localhost:3000" + returnTo
	
	// Set access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = session.AccessToken
	accessCookie.Expires = session.AccessExpiry
	accessCookie.Path = "/"
	accessCookie.HttpOnly = true
	accessCookie.Secure = true
	accessCookie.SameSite = http.SameSiteLaxMode // Changed from StrictMode for OAuth
	
	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = session.RefreshToken
	refreshCookie.Expires = session.RefreshExpiry
	refreshCookie.Path = "/"
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true
	refreshCookie.SameSite = http.SameSiteLaxMode // Changed from StrictMode for OAuth
	
	ctx.SetCookie(accessCookie)
	ctx.SetCookie(refreshCookie)
	
	// Redirect to frontend
	return ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
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
	
	// Set access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = session.AccessToken
	accessCookie.Expires = session.AccessExpiry
	accessCookie.Path = "/"
	accessCookie.HttpOnly = true
	accessCookie.Secure = true
	accessCookie.SameSite = http.SameSiteLaxMode // Changed from StrictMode for OAuth
	
	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = session.RefreshToken
	refreshCookie.Expires = session.RefreshExpiry
	refreshCookie.Path = "/"
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true
	refreshCookie.SameSite = http.SameSiteLaxMode // Changed from StrictMode for OAuth
	
	ctx.SetCookie(accessCookie)
	ctx.SetCookie(refreshCookie)
	
	// Redirect to frontend
	return ctx.Redirect(http.StatusTemporaryRedirect, "/")
}