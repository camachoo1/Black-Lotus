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
	// Setup
	ctx := context.Background()
	userID := uuid.New()
	tripID := uuid.New()

	t.Run("Trip Exists and Owned by User", func(t *testing.T) {
		// Setup expected trip
		expectedTrip := &models.Trip{
			ID:          tripID,
			UserID:      userID,
			Name:        "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(7 * 24 * time.Hour),
			Destination: "Test City",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock repository
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, tid uuid.UUID) (*models.Trip, error) {
				if tid != tripID {
					t.Errorf("Expected tripID %s, got %s", tripID, tid)
				}
				return expectedTrip, nil
			},
		}

		// Create service with mock repository
		tripService := services.NewTripService(mockRepo)

		// Call service
		trip, err := tripService.GetTripByID(ctx, tripID, userID)

		// Validate
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if trip != expectedTrip {
			t.Errorf("Expected trip %+v, got %+v", expectedTrip, trip)
		}
	})

	t.Run("Trip Not Found", func(t *testing.T) {
		// Mock repository
		mockRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, tid uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}

		// Create service with mock repository
		tripService := services.NewTripService(mockRepo)

		// Call service
		err := tripService.DeleteTrip(ctx, tripID, userID)

		// Validate
		if err == nil {
			t.Error("Expected error from repository, got nil")
		}

		if err != nil && err.Error() != "trip not found" {
			t.Errorf("Expected error message 'trip not found', got '%s'", err.Error())
		}
	})
}
