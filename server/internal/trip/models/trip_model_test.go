package models_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"black-lotus/internal/trip/models"
)

func TestTripValidation(t *testing.T) {
	validate := validator.New()

	// Test case: Valid trip
	t.Run("Valid Trip", func(t *testing.T) {
		now := time.Now()
		future := now.Add(7 * 24 * time.Hour)

		trip := models.Trip{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "Test Trip",
			Description: "A test trip",
			StartDate:   now,
			EndDate:     future,
			Destination: "Test City",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := validate.Struct(trip); err != nil {
			t.Errorf("Expected valid trip, got validation error: %v", err)
		}
	})

	// Test case: Missing required fields
	t.Run("Missing Required Fields", func(t *testing.T) {
		trip := models.Trip{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "Test Trip",
			Description: "A test trip",
			// Missing StartDate, EndDate, and Destination
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := validate.Struct(trip); err == nil {
			t.Error("Expected validation error for missing required fields, got nil")
		}
	})
}

func TestCreateTripInputValidation(t *testing.T) {
	validate := validator.New()

	// Test case: Valid input
	t.Run("Valid Input", func(t *testing.T) {
		now := time.Now()
		future := now.Add(7 * 24 * time.Hour)

		input := models.CreateTripInput{
			Name:        "Test Trip",
			Description: "A test trip",
			StartDate:   now,
			EndDate:     future,
			Destination: "Test City",
		}

		if err := validate.Struct(input); err != nil {
			t.Errorf("Expected valid input, got validation error: %v", err)
		}
	})

	// Test case: Missing required fields
	t.Run("Missing Required Fields", func(t *testing.T) {
		input := models.CreateTripInput{
			Name:        "Test Trip",
			Description: "A test trip",
			// Missing StartDate, EndDate, and Destination
		}

		if err := validate.Struct(input); err == nil {
			t.Error("Expected validation error for missing required fields, got nil")
		}
	})
}

func TestUpdateTripInput(t *testing.T) {
	// Test partial updates with UpdateTripInput
	t.Run("Partial Updates", func(t *testing.T) {
		namePtr := func(s string) *string { return &s }
		timePtr := func(tm time.Time) *time.Time { return &tm }

		now := time.Now()
		tomorrow := now.Add(24 * time.Hour)

		input := models.UpdateTripInput{
			Name:        namePtr("Updated Trip"),
			Description: namePtr("Updated description"),
			StartDate:   timePtr(now),
			EndDate:     timePtr(tomorrow),
			Destination: namePtr("Updated City"),
		}

		// No validation tags on UpdateTripInput, so we just check values
		if *input.Name != "Updated Trip" {
			t.Errorf("Expected name to be 'Updated Trip', got %s", *input.Name)
		}

		if !input.EndDate.After(*input.StartDate) {
			t.Error("Expected end date to be after start date")
		}
	})
}
