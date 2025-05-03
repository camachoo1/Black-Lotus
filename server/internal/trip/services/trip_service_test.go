package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/trip/models"
	"black-lotus/internal/trip/services"
)

// Mock TripRepository for testing
type MockTripRepository struct {
	CreateTripFn       func(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	GetTripByIDFn      func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
	UpdateTripFn       func(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	DeleteTripFn       func(ctx context.Context, tripID uuid.UUID) error
	GetTripsByUserIDFn func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
}

func (m *MockTripRepository) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	return m.CreateTripFn(ctx, userID, input)
}

func (m *MockTripRepository) GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	return m.GetTripByIDFn(ctx, tripID)
}

func (m *MockTripRepository) UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	return m.UpdateTripFn(ctx, tripID, input)
}

func (m *MockTripRepository) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	return m.DeleteTripFn(ctx, tripID)
}

func (m *MockTripRepository) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error) {
	return m.GetTripsByUserIDFn(ctx, userID, limit, offset)
}

// Helper functions for creating pointers - running into nil pointer dereferencing issues
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestCreateTrip(t *testing.T) {
	// Setup
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Valid Trip Creation", func(t *testing.T) {
		// Setup input and expected output
		input := models.CreateTripInput{
			Name:        "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(7 * 24 * time.Hour),
			Destination: "Test City",
		}

		expectedTrip := &models.Trip{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        input.Name,
			Description: input.Description,
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Destination: input.Destination,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock repository
		mockRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				if uid != userID {
					t.Errorf("Expected userID %s, got %s", userID, uid)
				}
				if inp.Name != input.Name {
					t.Errorf("Expected name %s, got %s", input.Name, inp.Name)
				}
				return expectedTrip, nil
			},
		}

		// Create service with mock repository
		tripService := services.NewTripService(mockRepo)

		// Call service
		trip, err := tripService.CreateTrip(ctx, userID, input)

		// Validate
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if trip != expectedTrip {
			t.Errorf("Expected trip %+v, got %+v", expectedTrip, trip)
		}
	})

	t.Run("Invalid Date Range", func(t *testing.T) {
		// Setup input with end date before start date
		input := models.CreateTripInput{
			Name:        "Invalid Trip",
			Description: "Test Description",
			StartDate:   time.Now().Add(7 * 24 * time.Hour), // 7 days in future
			EndDate:     time.Now().Add(24 * time.Hour),     // 1 day in future
			Destination: "Test City",
		}

		// Mock repository that should not be called
		mockRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				t.Error("Repository should not be called for invalid date range")
				return nil, nil
			},
		}

		// Create service with mock repository
		tripService := services.NewTripService(mockRepo)

		// Call service
		trip, err := tripService.CreateTrip(ctx, userID, input)

		// Validate
		if err == nil {
			t.Error("Expected error for invalid date range, got nil")
		}

		if err != nil && err.Error() != "end date cannot be before start date" {
			t.Errorf("Expected error message 'end date cannot be before start date', got '%s'", err.Error())
		}

		if trip != nil {
			t.Errorf("Expected nil trip, got %+v", trip)
		}
	})

	t.Run("Empty Name with Auto-generation", func(t *testing.T) {
		// Setup input with empty name
		input := models.CreateTripInput{
			Name:        "",
			Description: "Test Description",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(7 * 24 * time.Hour),
			Destination: "Paris",
		}

		expectedTrip := &models.Trip{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        "Trip to Paris",
			Description: input.Description,
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Destination: input.Destination,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock repository
		mockRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				if inp.Name != "Trip to Paris" {
					t.Errorf("Expected auto-generated name 'Trip to Paris', got '%s'", inp.Name)
				}
				return expectedTrip, nil
			},
		}

		// Create service with mock repository
		tripService := services.NewTripService(mockRepo)

		// Call service
		trip, err := tripService.CreateTrip(ctx, userID, input)

		// Validate
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if trip != expectedTrip {
			t.Errorf("Expected trip %+v, got %+v", expectedTrip, trip)
		}
	})

	t.Run("Repository Error", func(t *testing.T) {
		// Setup input
		input := models.CreateTripInput{
			Name:        "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(7 * 24 * time.Hour),
			Destination: "Test City",
		}

		// Mock repository
		mockRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				return nil, errors.New("database error")
			},
		}

		// Create service with mock repository
		tripService := services.NewTripService(mockRepo)

		// Call service
		trip, err := tripService.CreateTrip(ctx, userID, input)

		// Validate
		if err == nil {
			t.Error("Expected error from repository, got nil")
		}

		if err != nil && err.Error() != "database error" {
			t.Errorf("Expected error message 'database error', got '%s'", err.Error())
		}

		if trip != nil {
			t.Errorf("Expected nil trip, got %+v", trip)
		}
	})
}

func TestGetTripByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tripID := uuid.New()
	anotherUserID := uuid.New()

	t.Run("Trip Not Found", func(t *testing.T) {
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}

		service := services.NewTripService(mockRepo)

		// Call the service method
		trip, err := service.GetTripByID(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: anotherUserID, // Different user
				}, nil
			},
		}

		service := services.NewTripService(mockRepo)

		// Call the service method
		trip, err := service.GetTripByID(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected unauthorized error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})
}

func TestUpdateTrip(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tripID := uuid.New()
	anotherUserID := uuid.New()

	t.Run("Trip Not Found", func(t *testing.T) {
		// Setup mock repository
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}

		service := services.NewTripService(mockRepo)
		updateInput := models.UpdateTripInput{
			Name: stringPtr("Updated Trip"),
		}

		// Call the service method
		trip, err := service.UpdateTrip(ctx, tripID, userID, updateInput)

		// Verify results
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		// Setup mock repository
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      anotherUserID, // Different user
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   time.Now(),
					EndDate:     time.Now().Add(24 * time.Hour),
					Destination: "Original City",
				}, nil
			},
		}

		service := services.NewTripService(mockRepo)
		updateInput := models.UpdateTripInput{
			Name: stringPtr("Updated Trip"),
		}

		// Call the service method
		trip, err := service.UpdateTrip(ctx, tripID, userID, updateInput)

		// Verify results
		if err == nil {
			t.Error("Expected unauthorized error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Invalid Date Range - Both Dates", func(t *testing.T) {
		// Setup mock repository
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      userID, // Same user
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   time.Now(),
					EndDate:     time.Now().Add(24 * time.Hour),
					Destination: "Original City",
				}, nil
			},
		}

		service := services.NewTripService(mockRepo)

		futureDate := time.Now().Add(48 * time.Hour)
		pastDate := time.Now().Add(24 * time.Hour)
		updateInput := models.UpdateTripInput{
			StartDate: timePtr(futureDate),
			EndDate:   timePtr(pastDate), // Before start date
		}

		// Call the service method
		trip, err := service.UpdateTrip(ctx, tripID, userID, updateInput)

		// Verify results
		if err == nil {
			t.Error("Expected date validation error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Invalid Date Range - StartDate Only", func(t *testing.T) {
		// Setup mock repository
		now := time.Now()
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      userID,
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   now,
					EndDate:     now.Add(24 * time.Hour),
					Destination: "Original City",
				}, nil
			},
		}

		service := services.NewTripService(mockRepo)

		// New start date is after the existing end date
		newStartDate := now.Add(48 * time.Hour)
		updateInput := models.UpdateTripInput{
			StartDate: timePtr(newStartDate),
		}

		// Call the service method
		trip, err := service.UpdateTrip(ctx, tripID, userID, updateInput)

		// Verify results
		if err == nil {
			t.Error("Expected date validation error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Invalid Date Range - EndDate Only", func(t *testing.T) {
		// Setup mock repository
		now := time.Now()
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      userID,
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   now.Add(48 * time.Hour), // Future start
					EndDate:     now.Add(72 * time.Hour),
					Destination: "Original City",
				}, nil
			},
		}

		service := services.NewTripService(mockRepo)

		// New end date is before the existing start date
		newEndDate := now.Add(24 * time.Hour)
		updateInput := models.UpdateTripInput{
			EndDate: timePtr(newEndDate),
		}

		// Call the service method
		trip, err := service.UpdateTrip(ctx, tripID, userID, updateInput)

		// Verify results
		if err == nil {
			t.Error("Expected date validation error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Valid Update", func(t *testing.T) {
		// Setup mock repository
		now := time.Now()
		originalTrip := &models.Trip{
			ID:          tripID,
			UserID:      userID,
			Name:        "Original Trip",
			Description: "Original Description",
			StartDate:   now,
			EndDate:     now.Add(72 * time.Hour),
			Destination: "Original City",
		}

		updatedTrip := &models.Trip{
			ID:          tripID,
			UserID:      userID,
			Name:        "Updated Trip",
			Description: "Updated Description",
			StartDate:   now.Add(24 * time.Hour),
			EndDate:     now.Add(96 * time.Hour),
			Destination: "Updated City",
		}

		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return originalTrip, nil
			},
			UpdateTripFn: func(ctx context.Context, id uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
				return updatedTrip, nil
			},
		}

		service := services.NewTripService(mockRepo)

		updateInput := models.UpdateTripInput{
			Name:        stringPtr("Updated Trip"),
			Description: stringPtr("Updated Description"),
			StartDate:   timePtr(now.Add(24 * time.Hour)),
			EndDate:     timePtr(now.Add(96 * time.Hour)),
			Destination: stringPtr("Updated City"),
		}

		// Call the service method
		trip, err := service.UpdateTrip(ctx, tripID, userID, updateInput)

		// Verify results
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if trip == nil {
			t.Error("Expected trip, got nil")
		}
		if trip != nil && trip.Name != "Updated Trip" {
			t.Errorf("Expected updated name, got: %s", trip.Name)
		}
	})
}

// TestDeleteTrip tests the DeleteTrip method
func TestDeleteTrip(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tripID := uuid.New()
	anotherUserID := uuid.New()

	t.Run("Trip Not Found", func(t *testing.T) {
		// Already covered by existing test
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: anotherUserID, // Different user
				}, nil
			},
		}

		service := services.NewTripService(mockRepo)

		// Call the service method
		err := service.DeleteTrip(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected unauthorized error, got nil")
		}
	})

	t.Run("Successful Delete", func(t *testing.T) {
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: userID, // Same user
				}, nil
			},
			DeleteTripFn: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}

		service := services.NewTripService(mockRepo)

		// Call the service method
		err := service.DeleteTrip(ctx, tripID, userID)

		// Verify results
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}
