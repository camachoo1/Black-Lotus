package trips

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/session"
)

type Handler struct {
	service        ServiceInterface
	sessionService session.ServiceInterface
	validator      *validator.Validate
}

func NewHandler(service ServiceInterface, sessionService session.ServiceInterface) *Handler {
	validate := validator.New()

	// Register struct-level validation
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Handler{
		service:        service,
		sessionService: sessionService,
		validator:      validate,
	}
}

// CreateTrip creates a new trip for the authenticated user
func (h *Handler) CreateTrip(ctx echo.Context) error {
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

	// Parse request body
	var input models.CreateTripInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate the input
	if err := h.validator.Struct(input); err != nil {
		// Extract validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)

			for _, e := range validationErrors {
				switch e.Tag() {
				case "required":
					errorMessages[e.Field()] = fmt.Sprintf("%s is required", e.Field())
				default:
					errorMessages[e.Field()] = fmt.Sprintf("%s is invalid", e.Field())
				}
			}

			return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "Invalid request body",
				"details": errorMessages,
			})
		}

		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Create the trip
	trip, err := h.service.CreateTrip(ctx.Request().Context(), session.UserID, input)
	if err != nil {
		log.Printf("Failed to create trip: %v", err)

		// Handle specific business logic errors
		if err.Error() == "end date cannot be before start date" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
		}

		// For consistency with tests, return 500 for NonValidationError
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create trip",
		})
	}

	return ctx.JSON(http.StatusCreated, trip)
}

// GetTrip retrieves a specific trip by ID
func (h *Handler) GetTrip(ctx echo.Context) error {
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

	// Parse trip ID from URL
	tripID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid trip ID",
		})
	}

	// Get the trip
	trip, err := h.service.GetTripByID(ctx.Request().Context(), tripID, session.UserID)
	if err != nil {
		if err.Error() == "trip not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Trip not found",
			})
		}
		if err.Error() == "unauthorized access to trip" {
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "You do not have permission to view this trip",
			})
		}

		log.Printf("Failed to get trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get trip",
		})
	}

	return ctx.JSON(http.StatusOK, trip)
}

// GetUserTrips retrieves all trips for the authenticated user
func (h *Handler) GetUserTrips(ctx echo.Context) error {
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

	// Get the trips
	trips, err := h.service.GetTripsByUserID(ctx.Request().Context(), session.UserID, limit, offset)
	if err != nil {
		log.Printf("Failed to get trips: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get trips",
		})
	}

	return ctx.JSON(http.StatusOK, trips)
}

// UpdateTrip updates a specific trip by ID
func (h *Handler) UpdateTrip(ctx echo.Context) error {
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

	// Parse trip ID from URL
	tripID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid trip ID",
		})
	}

	// Parse request body
	var input models.UpdateTripInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Reject empty updates - add this check
	if input.Name == nil && input.Description == nil &&
		input.StartDate == nil && input.EndDate == nil &&
		input.Location == nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate the input
	if err := h.validator.Struct(input); err != nil {
		// Extract validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)

			for _, e := range validationErrors {
				errorMessages[e.Field()] = fmt.Sprintf("%s is invalid", e.Field())
			}

			return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "Invalid request body",
				"details": errorMessages,
			})
		}

		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Update the trip
	updatedTrip, err := h.service.UpdateTrip(ctx.Request().Context(), tripID, session.UserID, input)
	if err != nil {
		if err.Error() == "unauthorized access to trip" {
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "You do not have permission to update this trip",
			})
		} else if err.Error() == "trip not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Trip not found",
			})
		} else if err.Error() == "end date cannot be before start date" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
		}

		log.Printf("Failed to update trip: %v", err)
		// Always return BadRequest with consistent error message
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	return ctx.JSON(http.StatusOK, updatedTrip)
}

// DeleteTrip deletes a specific trip by ID
func (h *Handler) DeleteTrip(ctx echo.Context) error {
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

	// Parse trip ID from URL
	tripID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid trip ID",
		})
	}

	// Delete the trip
	err = h.service.DeleteTrip(ctx.Request().Context(), tripID, session.UserID)
	if err != nil {
		if err.Error() == "unauthorized access to trip" {
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "You do not have permission to delete this trip",
			})
		} else if err.Error() == "trip not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Trip not found",
			})
		}

		log.Printf("Failed to delete trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete trip",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Trip deleted successfully",
	})
}
