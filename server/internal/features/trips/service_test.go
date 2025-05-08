package trips_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/trips"
)

// MockRepository implements trips.Repository for testing
type MockRepository struct {
	createTripFunc       func(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	getTripByIDFunc      func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
	updateTripFunc       func(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	deleteTripFunc       func(ctx context.Context, tripID uuid.UUID) error
	getTripsByUserIDFunc func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
	getTripWithUserFunc  func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
}

func (m *MockRepository) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	if m.createTripFunc != nil {
		return m.createTripFunc(ctx, userID, input)
	}
	return nil, errors.New("CreateTrip not implemented")
}

func (m *MockRepository) GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	if m.getTripByIDFunc != nil {
		return m.getTripByIDFunc(ctx, tripID)
	}
	return nil, errors.New("GetTripByID not implemented")
}

func (m *MockRepository) UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	if m.updateTripFunc != nil {
		return m.updateTripFunc(ctx, tripID, input)
	}
	return nil, errors.New("UpdateTrip not implemented")
}

func (m *MockRepository) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	if m.deleteTripFunc != nil {
		return m.deleteTripFunc(ctx, tripID)
	}
	return errors.New("DeleteTrip not implemented")
}

func (m *MockRepository) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error) {
	if m.getTripsByUserIDFunc != nil {
		return m.getTripsByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetTripsByUserID not implemented")
}

func (m *MockRepository) GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	if m.getTripWithUserFunc != nil {
		return m.getTripWithUserFunc(ctx, tripID)
	}
	return nil, errors.New("GetTripWithUser not implemented")
}

// MockViewService implements the view.ServiceInterface for testing
type MockViewService struct {
	getUserProfileFunc func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *MockViewService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, userID)
	}
	return nil, errors.New("GetUserProfile not implemented")
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper function to setup service for testing
func setupServiceTest() (trips.ServiceInterface, *MockRepository, *MockViewService) {
	mockRepo := &MockRepository{}
	mockViewService := &MockViewService{}
	service := trips.NewService(mockRepo, mockViewService)
	return service, mockRepo, mockViewService
}

func TestServiceCreateTrip(t *testing.T) {
	testCases := []struct {
		name          string
		input         models.CreateTripInput
		setupMocks    func(*testing.T, *MockRepository, *MockViewService)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulCreation",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService) {
				mockRepo.createTripFunc = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
					return &models.Trip{
						ID:          uuid.New(),
						UserID:      uid,
						Name:        inp.Name,
						Description: inp.Description,
						StartDate:   inp.StartDate,
						EndDate:     inp.EndDate,
						Location:    inp.Location,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "InvalidDateRange",
			input: models.CreateTripInput{
				Name:        "Invalid Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(7 * 24 * time.Hour), // 7 days future
				EndDate:     time.Now().Add(24 * time.Hour),     // 1 day future
				Location:    "Test City",
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService) {
				// Repository should not be called
			},
			expectedError: true,
			errorMessage:  "end date cannot be before start date",
		},
		{
			name: "EmptyNameAutoGeneration",
			input: models.CreateTripInput{
				Name:        "",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Paris",
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService) {
				mockRepo.createTripFunc = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
					if inp.Name != "Trip to Paris" {
						t.Errorf("Expected auto-generated name 'Trip to Paris', got '%s'", inp.Name)
					}
					return &models.Trip{
						ID:          uuid.New(),
						UserID:      uid,
						Name:        inp.Name,
						Description: inp.Description,
						StartDate:   inp.StartDate,
						EndDate:     inp.EndDate,
						Location:    inp.Location,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "RepositoryError",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService) {
				mockRepo.createTripFunc = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: true,
			errorMessage:  "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService)

			// Execute
			result, err := service.CreateTrip(context.Background(), userID, tc.input)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected trip to be returned, got nil")
				}

				if tc.input.Name == "" {
					if result.Name != "Trip to "+tc.input.Location {
						t.Errorf("Expected name 'Trip to %s', got '%s'", tc.input.Location, result.Name)
					}
				} else if result.Name != tc.input.Name {
					t.Errorf("Expected name '%s', got '%s'", tc.input.Name, result.Name)
				}
			}
		})
	}
}

func TestGetTripByID(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*testing.T, *MockRepository, *MockViewService, uuid.UUID, uuid.UUID)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulRetrieval",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: userID,
						Name:   "Test Trip",
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "TripNotFound",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return nil, errors.New("trip not found")
				}
			},
			expectedError: true,
			errorMessage:  "trip not found",
		},
		{
			name: "UnauthorizedAccess",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: uuid.New(), // Different user ID
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "unauthorized access to trip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			tripID := uuid.New()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService, tripID, userID)

			// Execute
			result, err := service.GetTripByID(context.Background(), tripID, userID)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected trip to be returned, got nil")
				}
				if result.ID != tripID {
					t.Errorf("Expected trip ID %s, got %s", tripID, result.ID)
				}
				if result.UserID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, result.UserID)
				}
			}
		})
	}
}

func TestGetTripWithUser(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*testing.T, *MockRepository, *MockViewService, uuid.UUID, uuid.UUID)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulRetrieval",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripWithUserFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: userID,
						Name:   "Test Trip",
						User: &models.User{
							ID:   userID,
							Name: "Test User",
						},
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "TripNotFound",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripWithUserFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return nil, errors.New("trip not found")
				}
			},
			expectedError: true,
			errorMessage:  "trip not found",
		},
		{
			name: "UnauthorizedAccess",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripWithUserFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: uuid.New(), // Different user ID
						User: &models.User{
							ID:   uuid.New(),
							Name: "Another User",
						},
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "unauthorized access to trip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			tripID := uuid.New()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService, tripID, userID)

			// Execute
			result, err := service.GetTripWithUser(context.Background(), tripID, userID)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected trip to be returned, got nil")
				}
				if result.User == nil {
					t.Fatal("Expected user data to be included, got nil")
				}
			}
		})
	}
}

func TestGetTripsByUserID(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*testing.T, *MockRepository, *MockViewService, uuid.UUID)
		expectedError bool
		errorMessage  string
		tripCount     int
	}{
		{
			name: "SuccessfulRetrieval",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return &models.User{
						ID:   userID,
						Name: "Test User",
					}, nil
				}

				mockRepo.getTripsByUserIDFunc = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
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
				}
			},
			expectedError: false,
			tripCount:     2,
		},
		{
			name: "UserProfileError",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("failed to get user profile")
				}
			},
			expectedError: true,
			errorMessage:  "failed to get user profile",
			tripCount:     0,
		},
		{
			name: "NilUserReturned",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, nil
				}
			},
			expectedError: true,
			errorMessage:  "user not found",
			tripCount:     0,
		},
		{
			name: "RepositoryError",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return &models.User{
						ID:   userID,
						Name: "Test User",
					}, nil
				}

				mockRepo.getTripsByUserIDFunc = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: true,
			errorMessage:  "database error",
			tripCount:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService, userID)

			// Execute
			result, err := service.GetTripsByUserID(context.Background(), userID, 10, 0)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if len(result) != tc.tripCount {
					t.Errorf("Expected %d trips, got %d", tc.tripCount, len(result))
				}
			}
		})
	}
}

func TestServiceUpdateTrip(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name          string
		updateInput   models.UpdateTripInput
		setupMocks    func(*testing.T, *MockRepository, *MockViewService, uuid.UUID, uuid.UUID)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulUpdate",
			updateInput: models.UpdateTripInput{
				Name:        stringPtr("Updated Trip"),
				Description: stringPtr("Updated Description"),
				StartDate:   timePtr(now.Add(24 * time.Hour)),
				EndDate:     timePtr(now.Add(96 * time.Hour)),
				Location:    stringPtr("Updated City"),
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:        tripID,
						UserID:    userID,
						Name:      "Original Trip",
						StartDate: now,
						EndDate:   now.Add(72 * time.Hour),
					}, nil
				}
				mockRepo.updateTripFunc = func(ctx context.Context, id uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					return &models.Trip{
						ID:          tripID,
						UserID:      userID,
						Name:        *input.Name,
						Description: *input.Description,
						StartDate:   *input.StartDate,
						EndDate:     *input.EndDate,
						Location:    *input.Location,
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "TripNotFound",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return nil, errors.New("trip not found")
				}
			},
			expectedError: true,
			errorMessage:  "trip not found",
		},
		{
			name: "UnauthorizedAccess",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: uuid.New(), // Different user ID
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "unauthorized access to trip",
		},
		{
			name: "InvalidDateRange",
			updateInput: models.UpdateTripInput{
				StartDate: timePtr(now.Add(48 * time.Hour)),
				EndDate:   timePtr(now.Add(24 * time.Hour)),
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:        tripID,
						UserID:    userID,
						StartDate: now,
						EndDate:   now.Add(24 * time.Hour),
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "end date cannot be before start date",
		},
		{
			name: "InvalidDateRange_StartDateOnly",
			updateInput: models.UpdateTripInput{
				StartDate: timePtr(now.Add(48 * time.Hour)),
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:        tripID,
						UserID:    userID,
						StartDate: now,
						EndDate:   now.Add(24 * time.Hour),
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "end date cannot be before start date",
		},
		{
			name: "InvalidDateRange_EndDateOnly",
			updateInput: models.UpdateTripInput{
				EndDate: timePtr(now.Add(-24 * time.Hour)), // Before trip.StartDate
			},
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:        tripID,
						UserID:    userID,
						StartDate: now,
						EndDate:   now.Add(72 * time.Hour),
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "end date cannot be before start date",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			tripID := uuid.New()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService, tripID, userID)

			// Execute
			result, err := service.UpdateTrip(context.Background(), tripID, userID, tc.updateInput)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected trip to be returned, got nil")
				}
				if tc.updateInput.Name != nil && result.Name != *tc.updateInput.Name {
					t.Errorf("Expected name '%s', got '%s'", *tc.updateInput.Name, result.Name)
				}
			}
		})
	}
}

func TestServiceDeleteTrip(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*testing.T, *MockRepository, *MockViewService, uuid.UUID, uuid.UUID)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulDelete",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: userID,
					}, nil
				}
				mockRepo.deleteTripFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			expectedError: false,
		},
		{
			name: "TripNotFound",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return nil, errors.New("trip not found")
				}
			},
			expectedError: true,
			errorMessage:  "trip not found",
		},
		{
			name: "UnauthorizedAccess",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, tripID, userID uuid.UUID) {
				mockRepo.getTripByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
					return &models.Trip{
						ID:     tripID,
						UserID: uuid.New(), // Different user ID
					}, nil
				}
			},
			expectedError: true,
			errorMessage:  "unauthorized access to trip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			tripID := uuid.New()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService, tripID, userID)

			// Execute
			err := service.DeleteTrip(context.Background(), tripID, userID)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestServiceGetUserWithTrips(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*testing.T, *MockRepository, *MockViewService, uuid.UUID)
		expectedError bool
		errorMessage  string
		tripCount     int
	}{
		{
			name: "SuccessfulRetrieval",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return &models.User{
						ID:   userID,
						Name: "Test User",
					}, nil
				}
				mockRepo.getTripsByUserIDFunc = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
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
				}
			},
			expectedError: false,
			tripCount:     2,
		},
		{
			name: "UserNotFound",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("user not found")
				}
			},
			expectedError: true,
			errorMessage:  "user not found",
			tripCount:     0,
		},
		{
			name: "ErrorGettingTrips",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return &models.User{
						ID:   userID,
						Name: "Test User",
					}, nil
				}
				mockRepo.getTripsByUserIDFunc = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: true,
			errorMessage:  "database error",
			tripCount:     0,
		},
		{
			name: "NilUserReturned",
			setupMocks: func(t *testing.T, mockRepo *MockRepository, mockViewService *MockViewService, userID uuid.UUID) {
				mockViewService.getUserProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, nil
				}
			},
			expectedError: false,
			tripCount:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo, mockViewService := setupServiceTest()
			userID := uuid.New()

			// Setup mocks
			tc.setupMocks(t, mockRepo, mockViewService, userID)

			// Execute
			result, err := service.GetUserWithTrips(context.Background(), userID, 10, 0)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Special case for nil user
				if tc.name == "NilUserReturned" {
					if result != nil {
						t.Error("Expected nil user, got non-nil")
					}
					return
				}

				if result == nil {
					t.Fatal("Expected user to be returned, got nil")
				}

				if len(result.Trips) != tc.tripCount {
					t.Errorf("Expected %d trips, got %d", tc.tripCount, len(result.Trips))
				}
			}
		})
	}
}
