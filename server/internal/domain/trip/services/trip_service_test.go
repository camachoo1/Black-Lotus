package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/domain/trip/services"
	"black-lotus/internal/models"
)

// Updated MockTripRepository to include new relationship method
type MockTripRepository struct {
	CreateTripFn       func(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	GetTripByIDFn      func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
	UpdateTripFn       func(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	DeleteTripFn       func(ctx context.Context, tripID uuid.UUID) error
	GetTripsByUserIDFn func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
	GetTripWithUserFn  func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
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

// New method implementation
func (m *MockTripRepository) GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	if m.GetTripWithUserFn != nil {
		return m.GetTripWithUserFn(ctx, tripID)
	}
	return nil, errors.New("GetTripWithUser not implemented")
}

type MockUserRepository struct {
	GetUserByIDFn func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented")
}

// Implement required interface methods
func (m *MockUserRepository) CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
	return nil, errors.New("CreateUser not implemented")
}

func (m *MockUserRepository) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	return nil, errors.New("LoginUser not implemented")
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, errors.New("GetUserByEmail not implemented")
}

func (m *MockUserRepository) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error) {
	return nil, errors.New("GetUserWithTrips not implemented")
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Add test for the new GetTripWithUser method
func TestGetTripWithUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tripID := uuid.New()
	anotherUserID := uuid.New()

	t.Run("Trip Not Found", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{
			GetTripWithUserFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		trip, err := service.GetTripWithUser(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{
			GetTripWithUserFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: anotherUserID, // Different user
					User: &models.User{
						ID:   anotherUserID,
						Name: "Another User",
					},
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		trip, err := service.GetTripWithUser(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected unauthorized error, got nil")
		}
		if trip != nil {
			t.Errorf("Expected nil trip, got: %v", trip)
		}
	})

	t.Run("Successful Retrieval", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{
			GetTripWithUserFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: userID,
					Name:   "Test Trip",
					User: &models.User{
						ID:   userID,
						Name: "Test User",
					},
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		trip, err := service.GetTripWithUser(ctx, tripID, userID)

		// Verify results
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if trip == nil {
			t.Fatal("Expected trip, got nil")
		}
		if trip.User == nil {
			t.Fatal("Expected user data to be included, got nil")
		}
		if trip.User.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, trip.User.ID)
		}
	})
}

// Test for GetUserWithTrips
func TestGetUserWithTrips(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("User Not Found", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{}
		mockUserRepo := &MockUserRepository{
			GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("user not found")
			},
		}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		user, err := service.GetUserWithTrips(ctx, userID, 10, 0)

		// Verify results
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}
	})

	t.Run("Successful Retrieval", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{
			GetTripsByUserIDFn: func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
				return []*models.Trip{
					{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "Trip 1",
					},
					{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "Trip 2",
					},
				}, nil
			},
		}

		mockUserRepo := &MockUserRepository{
			GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:   userID,
					Name: "Test User",
				}, nil
			},
		}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		user, err := service.GetUserWithTrips(ctx, userID, 10, 0)

		// Verify results
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if user == nil {
			t.Fatal("Expected user, got nil")
		}
		if len(user.Trips) != 2 {
			t.Errorf("Expected 2 trips, got %d", len(user.Trips))
		}
	})
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
			Location:    "Test City",
		}

		expectedTrip := &models.Trip{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        input.Name,
			Description: input.Description,
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Location:    input.Location,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock repositories
		mockTripRepo := &MockTripRepository{
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
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		tripService := services.NewTripService(mockTripRepo, mockUserRepo)

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
			Location:    "Test City",
		}

		// Mock repository that should not be called
		mockTripRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				t.Error("Repository should not be called for invalid date range")
				return nil, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		tripService := services.NewTripService(mockTripRepo, mockUserRepo)

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
			Location:    "Paris",
		}

		expectedTrip := &models.Trip{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        "Trip to Paris",
			Description: input.Description,
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Location:    input.Location,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock repository
		mockTripRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				if inp.Name != "Trip to Paris" {
					t.Errorf("Expected auto-generated name 'Trip to Paris', got '%s'", inp.Name)
				}
				return expectedTrip, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		tripService := services.NewTripService(mockTripRepo, mockUserRepo)

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
			Location:    "Test City",
		}

		// Mock repository
		mockTripRepo := &MockTripRepository{
			CreateTripFn: func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
				return nil, errors.New("database error")
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		tripService := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: anotherUserID, // Different user
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      anotherUserID, // Different user
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   time.Now(),
					EndDate:     time.Now().Add(24 * time.Hour),
					Location:    "Original City",
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      userID, // Same user
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   time.Now(),
					EndDate:     time.Now().Add(24 * time.Hour),
					Location:    "Original City",
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      userID,
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   now,
					EndDate:     now.Add(24 * time.Hour),
					Location:    "Original City",
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:          tripID,
					UserID:      userID,
					Name:        "Original Trip",
					Description: "Original Description",
					StartDate:   now.Add(48 * time.Hour), // Future start
					EndDate:     now.Add(72 * time.Hour),
					Location:    "Original City",
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

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
			Location:    "Original City",
		}

		updatedTrip := &models.Trip{
			ID:          tripID,
			UserID:      userID,
			Name:        "Updated Trip",
			Description: "Updated Description",
			StartDate:   now.Add(24 * time.Hour),
			EndDate:     now.Add(96 * time.Hour),
			Location:    "Updated City",
		}

		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return originalTrip, nil
			},
			UpdateTripFn: func(ctx context.Context, id uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
				return updatedTrip, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		updateInput := models.UpdateTripInput{
			Name:        stringPtr("Updated Trip"),
			Description: stringPtr("Updated Description"),
			StartDate:   timePtr(now.Add(24 * time.Hour)),
			EndDate:     timePtr(now.Add(96 * time.Hour)),
			Location:    stringPtr("Updated City"),
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
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return nil, errors.New("trip not found")
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		err := service.DeleteTrip(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{
			GetTripByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
				return &models.Trip{
					ID:     tripID,
					UserID: anotherUserID, // Different user
				}, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		err := service.DeleteTrip(ctx, tripID, userID)

		// Verify results
		if err == nil {
			t.Error("Expected unauthorized error, got nil")
		}
	})

	t.Run("Successful Delete", func(t *testing.T) {
		mockTripRepo := &MockTripRepository{
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
		mockUserRepo := &MockUserRepository{}

		// Updated constructor call with both repositories
		service := services.NewTripService(mockTripRepo, mockUserRepo)

		// Call the service method
		err := service.DeleteTrip(ctx, tripID, userID)

		// Verify results
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}
