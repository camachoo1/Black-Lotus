package controllers

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

	"black-lotus/internal/domain/auth/services"
	"black-lotus/internal/domain/auth/validators"
	tripService "black-lotus/internal/domain/trip/services"
	"black-lotus/internal/models"
)

type TripController struct {
	tripService    *tripService.TripService
	sessionService *services.SessionService
	validator      *validator.Validate
}

func NewTripController(tripService *tripService.TripService, sessionService *services.SessionService) *TripController {
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

	return &TripController{
		tripService:    tripService,
		sessionService: sessionService,
		validator:      validate,
	}
}

// CreateTrip creates a new trip for the authenticated user
func (c *TripController) CreateTrip(ctx echo.Context) error {
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

	// Parse request body
	var input models.CreateTripInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate the input
	if err := c.validator.Struct(input); err != nil {
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
				"error":   "Validation failed",
				"details": errorMessages,
			})
		}

		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Create the trip
	trip, err := c.tripService.CreateTrip(ctx.Request().Context(), session.UserID, input)
	if err != nil {
		log.Printf("Failed to create trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create trip",
		})
	}

	return ctx.JSON(http.StatusCreated, trip)
}

// GetTrip retrieves a specific trip by ID
func (c *TripController) GetTrip(ctx echo.Context) error {
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

	// Parse trip ID from URL
	tripID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid trip ID",
		})
	}

	// Get the trip
	trip, err := c.tripService.GetTripByID(ctx.Request().Context(), tripID, session.UserID)
	if err != nil {
		if err.Error() == "trip not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Trip not found",
			})
		}

		log.Printf("Failed to get trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get trip",
		})
	}

	// Check if the trip belongs to the authenticated user
	if trip.UserID != session.UserID {
		return ctx.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to view this trip",
		})
	}

	return ctx.JSON(http.StatusOK, trip)
}

// GetUserTrips retrieves all trips for the authenticated user
func (c *TripController) GetUserTrips(ctx echo.Context) error {
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

	// Parse pagination parameters
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	offset, _ := strconv.Atoi(ctx.QueryParam("offset"))

	// Get the trips
	trips, err := c.tripService.GetTripsByUserID(ctx.Request().Context(), session.UserID, limit, offset)
	if err != nil {
		log.Printf("Failed to get trips: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get trips",
		})
	}

	return ctx.JSON(http.StatusOK, trips)
}

// UpdateTrip updates a specific trip by ID
func (c *TripController) UpdateTrip(ctx echo.Context) error {
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

	// Parse trip ID from URL
	tripID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid trip ID",
		})
	}

	// First get the trip to verify ownership
	trip, err := c.tripService.GetTripByID(ctx.Request().Context(), tripID, session.UserID)
	if err != nil {
		if err.Error() == "trip not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Trip not found",
			})
		}

		log.Printf("Failed to get trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get trip",
		})
	}

	// Check if the trip belongs to the authenticated user
	if trip.UserID != session.UserID {
		return ctx.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to update this trip",
		})
	}

	// Parse request body
	var input models.UpdateTripInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate the input
	if err := c.validator.Struct(input); err != nil {
		// Extract validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)

			for _, e := range validationErrors {
				errorMessages[e.Field()] = fmt.Sprintf("%s is invalid", e.Field())
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

	// Update the trip
	updatedTrip, err := c.tripService.UpdateTrip(ctx.Request().Context(), tripID, session.UserID, input)
	if err != nil {
		log.Printf("Failed to update trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update trip",
		})
	}

	return ctx.JSON(http.StatusOK, updatedTrip)
}

// DeleteTrip deletes a specific trip by ID
func (c *TripController) DeleteTrip(ctx echo.Context) error {
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

	// Parse trip ID from URL
	tripID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid trip ID",
		})
	}

	// First get the trip to verify ownership
	trip, err := c.tripService.GetTripByID(ctx.Request().Context(), tripID, session.UserID)
	if err != nil {
		if err.Error() == "trip not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Trip not found",
			})
		}

		log.Printf("Failed to get trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get trip",
		})
	}

	// Check if the trip belongs to the authenticated user
	if trip.UserID != session.UserID {
		return ctx.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to delete this trip",
		})
	}

	// Delete the trip
	err = c.tripService.DeleteTrip(ctx.Request().Context(), tripID, session.UserID)
	if err != nil {
		log.Printf("Failed to delete trip: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete trip",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Trip deleted successfully",
	})
}
