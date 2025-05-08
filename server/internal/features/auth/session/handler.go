package session

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RefreshToken(ctx echo.Context) error {
	// Get refresh token from cookie
	refreshCookie, err := ctx.Cookie("refresh_token")
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No refresh token provided",
		})
	}

	// Use the refresh token to get a new access token
	session, err := h.service.RefreshAccessToken(ctx.Request().Context(), refreshCookie.Value)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid refresh token",
		})
	}

	// Set the new access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = session.AccessToken
	accessCookie.Expires = session.AccessExpiry
	accessCookie.Path = "/"
	accessCookie.HttpOnly = true

	// For production
	accessCookie.Secure = true
	accessCookie.SameSite = http.SameSiteLaxMode

	ctx.SetCookie(accessCookie)

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Access token refreshed successfully",
	})
}

func (h *Handler) LogoutUser(ctx echo.Context) error {
	// Try to get both tokens
	accessCookie, accessErr := ctx.Cookie("access_token")
	refreshCookie, refreshErr := ctx.Cookie("refresh_token")

	// Check if already logged out
	if accessErr != nil && refreshErr != nil {
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Already logged out",
		})
	}

	// Delete session by access token if it exists
	if accessErr == nil {
		err := h.service.EndSessionByAccessToken(ctx.Request().Context(), accessCookie.Value)
		if err != nil {
			// Log the error but continue
			log.Printf("Failed to end session by access token: %v", err)
		}
	}

	// Delete session by refresh token if it exists
	if refreshErr == nil {
		err := h.service.EndSessionByRefreshToken(ctx.Request().Context(), refreshCookie.Value)
		if err != nil {
			// Log the error but continue
			log.Printf("Failed to end session by refresh token: %v", err)
		}
	}

	// Clear access token cookie
	accessCookieClear := new(http.Cookie)
	accessCookieClear.Name = "access_token"
	accessCookieClear.Value = ""
	accessCookieClear.MaxAge = -1 // Expire immediately
	accessCookieClear.Path = "/"
	ctx.SetCookie(accessCookieClear)

	// Clear refresh token cookie
	refreshCookieClear := new(http.Cookie)
	refreshCookieClear.Name = "refresh_token"
	refreshCookieClear.Value = ""
	refreshCookieClear.MaxAge = -1 // Expire immediately
	refreshCookieClear.Path = "/"
	ctx.SetCookie(refreshCookieClear)

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

func (h *Handler) GetCSRFToken(ctx echo.Context) error {
	token := ctx.Get("csrf").(string)

	return ctx.JSON(http.StatusOK, map[string]string{
		"csrf_token": token,
	})
}
