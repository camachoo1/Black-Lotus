package trips

import (
	"black-lotus/internal/features/auth/session"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service        ServiceInterface
	sessionService session.ServiceInterface
}

func NewHandler(service ServiceInterface, sessionService session.ServiceInterface) *Handler {
	return &Handler{
		service:        service,
		sessionService: sessionService,
	}
}

func (h *Handler) GetUserProfileWithTrips(ctx echo.Context) error {
	// Get access token from cookie
	accessCookie, err := ctx.Cookie("access_token")
	if err != nil {
		// No access token - check if there's a refresh token
		_, refreshErr := ctx.Cookie("refresh_token")
		if refreshErr != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Not authenticated",
			})
		}
		// Has refresh token but no access token - client should refresh
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Access token expired",
			"code":  "token_expired",
		})
	}

	// Validate access token
	session, err := h.sessionService.ValidateAccessToken(ctx.Request().Context(), accessCookie.Value)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid access token",
			"code":  "token_invalid",
		})
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	offset, _ := strconv.Atoi(ctx.QueryParam("offset"))

	user, err := h.service.GetUserWithTrips(ctx.Request().Context(), session.UserID, limit, offset)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user profile with trips",
		})
	}

	// Check if user is nil
	if user == nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return ctx.JSON(http.StatusOK, user)
}
