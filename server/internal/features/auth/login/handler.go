package login

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/session"
)

type Handler struct {
	service        *Service
	sessionService session.ServiceInterface
	validator      *validator.Validate
}

func NewHandler(service *Service, sessionService session.ServiceInterface, validator *validator.Validate) *Handler {
	return &Handler{
		service:        service,
		sessionService: sessionService,
		validator:      validator,
	}
}

func (h *Handler) Login(ctx echo.Context) error {
	var input models.LoginUserInput

	// Validate request data
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Authenticate user credentials
	user, err := h.service.LoginUser(ctx.Request().Context(), input)
	if err != nil {
		// Generic error for security (don't reveal if email or password was wrong)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials. Please check your email and password and try again.",
		})
	}

	// Create a session for the authenticated user
	session, err := h.sessionService.CreateSession(ctx.Request().Context(), user.ID)
	if err != nil {
		log.Printf("Session creation error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create session: " + err.Error(),
		})
	}

	// Set access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = session.AccessToken
	accessCookie.Expires = session.AccessExpiry
	accessCookie.Path = "/"
	accessCookie.HttpOnly = true
	// For production
	accessCookie.Secure = true
	accessCookie.SameSite = http.SameSiteStrictMode

	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = session.RefreshToken
	refreshCookie.Expires = session.RefreshExpiry
	refreshCookie.Path = "/"
	refreshCookie.HttpOnly = true
	// For production
	refreshCookie.Secure = true
	refreshCookie.SameSite = http.SameSiteStrictMode

	ctx.SetCookie(accessCookie)
	ctx.SetCookie(refreshCookie)

	return ctx.JSON(http.StatusOK, user)
}
