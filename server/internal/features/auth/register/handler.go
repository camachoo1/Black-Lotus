package register

import (
	"fmt"
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

func (h *Handler) Register(ctx echo.Context) error {
	var input models.CreateUserInput

	// Validate request data
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(input); err != nil {
		// Extract validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				switch e.Tag() {
				case "required":
					errorMessages[e.Field()] = fmt.Sprintf("%s is required", e.Field())
				case "email":
					errorMessages[e.Field()] = "Please enter a valid email address"
				case "min":
					errorMessages[e.Field()] = fmt.Sprintf("%s must be at least %s characters long", e.Field(), e.Param())
				case "containsuppercase":
					errorMessages[e.Field()] = "Password must contain at least one uppercase letter"
				case "containslowercase":
					errorMessages[e.Field()] = "Password must contain at least one lowercase letter"
				case "containsnumber":
					errorMessages[e.Field()] = "Password must contain at least one number"
				case "containsspecialchar":
					errorMessages[e.Field()] = "Password must contain at least one special character"
				default:
					errorMessages[e.Field()] = fmt.Sprintf("%s is invalid", e.Field())
				}
			}
			return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "Validation failed",
				"details": errorMessages,
			})
		}

		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Create the user
	user, err := h.service.Register(ctx.Request().Context(), input)
	if err != nil {
		// Check for specific errors
		if err.Error() == "user with this email already exists" {
			return ctx.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	// Create a session to automatically log in the new user
	session, err := h.sessionService.CreateSession(ctx.Request().Context(), user.ID)
	if err != nil {
		// User was created, but session creation failed
		// We'll still return success but log the error
		log.Printf("Failed to create session for new user: %v", err)
	} else {
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
	}

	return ctx.JSON(http.StatusCreated, user)
}
