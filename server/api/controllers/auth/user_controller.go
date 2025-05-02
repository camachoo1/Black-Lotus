package controllers

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/auth/models"
	"black-lotus/internal/auth/services"
	"black-lotus/internal/auth/validators"
)

type UserController struct {
	userService    *services.UserService
	sessionService *services.SessionService
	validator      *validator.Validate
}

func NewUserController(userService *services.UserService, sessionService *services.SessionService) *UserController {
	validate := validator.New()

	// This is critical - register struct-level validation
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	validators.RegisterCustomValidators(validate)

	return &UserController{
		userService:    userService,
		sessionService: sessionService,
		validator:      validate,
	}
}

// Creates a new user account and logs them in automatically
func (c *UserController) RegisterUser(ctx echo.Context) error {
	var input models.CreateUserInput

	// Validate request data
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.validator.Struct(input); err != nil {
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
	user, err := c.userService.CreateUser(ctx.Request().Context(), input)
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
	session, err := c.sessionService.CreateSession(ctx.Request().Context(), user.ID)
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

// LoginUser authenticates a user and creates a session
func (c *UserController) LoginUser(ctx echo.Context) error {
	var input models.LoginUserInput

	// Validate request data
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.validator.Struct(input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Authenticate user credentials
	user, err := c.userService.LoginUser(ctx.Request().Context(), input)
	if err != nil {
		// Generic error for security (don't reveal if email or password was wrong)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials. Please check your email and password and try again.",
		})
	}

	// Create a session for the authenticated user
	session, err := c.sessionService.CreateSession(ctx.Request().Context(), user.ID)
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

// LogoutUser ends the current user session
func (c *UserController) LogoutUser(ctx echo.Context) error {
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
		err := c.sessionService.EndSessionByAccessToken(ctx.Request().Context(), accessCookie.Value)
		if err != nil {
			// Log the error but continue
			log.Printf("Failed to end session by access token: %v", err)
		}
	}

	// Delete session by refresh token if it exists
	if refreshErr == nil {
		err := c.sessionService.EndSessionByRefreshToken(ctx.Request().Context(), refreshCookie.Value)
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

func (c *UserController) RefreshToken(ctx echo.Context) error {
	// Get refresh token from cookie
	refreshCookie, err := ctx.Cookie("refresh_token")
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No refresh token provided",
		})
	}

	// Use the refresh token to get a new access token
	session, err := c.sessionService.RefreshAccessToken(ctx.Request().Context(), refreshCookie.Value)
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

func (c *UserController) GetUserProfile(ctx echo.Context) error {
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
	session, err := c.sessionService.ValidateAccessToken(ctx.Request().Context(), accessCookie.Value)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid access token",
			"code":  "token_invalid",
		})
	}

	// Get user from session
	user, err := c.userService.GetUserByID(ctx.Request().Context(), session.UserID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user",
		})
	}

	return ctx.JSON(http.StatusOK, user)
}

func (c *UserController) GetCSRFToken(ctx echo.Context) error {
	token := ctx.Get("csrf").(string)

	return ctx.JSON(http.StatusOK, map[string]string{
		"csrf_token": token,
	})
}
