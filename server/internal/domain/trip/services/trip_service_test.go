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

// MockTripRepository implements all required repository methods
type MockTripRepository struct {
	CreateTripFn       func(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	GetTripByIDFn      func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
	UpdateTripFn       func(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	DeleteTripFn       func(ctx context.Context, tripID uuid.UUID) error
	GetTripsByUserIDFn func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
	GetTripWithUserFn  func(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
}

func (m *MockTripRepository) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	if m.CreateTripFn != nil {
		return m.CreateTripFn(ctx, userID, input)
	}
	return nil, errors.New("CreateTrip not implemented")
}

func (m *MockTripRepository) GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	if m.GetTripByIDFn != nil {
		return m.GetTripByIDFn(ctx, tripID)
	}
	return nil, errors.New("GetTripByID not implemented")
}

func (m *MockTripRepository) UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	if m.UpdateTripFn != nil {
		return m.UpdateTripFn(ctx, tripID, input)
	}
	return nil, errors.New("UpdateTrip not implemented")
}

func (m *MockTripRepository) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	if m.DeleteTripFn != nil {
		return m.DeleteTripFn(ctx, tripID)
	}
	return errors.New("DeleteTrip not implemented")
}

func (m *MockTripRepository) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error) {
	if m.GetTripsByUserIDFn != nil {
		return m.GetTripsByUserIDFn(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetTripsByUserID not implemented")
}

func (m *MockTripRepository) GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	if m.GetTripWithUserFn != nil {
		return m.GetTripWithUserFn(ctx, tripID)
	}
	return nil, errors.New("GetTripWithUser not implemented")
}

// MockUserRepository implements the minimal required methods
type MockUserRepository struct {
	GetUserByIDFn func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented")
}

// Implement other required interface methods as stubs
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

func (m *MockUserRepository) SetEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	return errors.New("SetEmailVerified not implemented")
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// setupTestService creates a new TripService with mock repositories
func setupTestService() (*services.TripService, *MockTripRepository, *MockUserRepository) {
	mockTripRepo := &MockTripRepository{}
	mockUserRepo := &MockUserRepository{}
	service := services.NewTripService(mockTripRepo, mockUserRepo)
	return service, mockTripRepo, mockUserRepo
}

func TestTripService(t *testing.T) {
	t.Run("CreateTrip", func(t *testing.T) {
		// Table-driven test for trip creation
		testCases := []struct {
			name          string
			input         models.CreateTripInput
			setupMocks    func(*MockTripRepository)
			expectedError bool
			errorMessage  string
			checkTrip     func(*testing.T, *models.Trip)
		}{
			{
				name: "ValidTripCreation",
				input: models.CreateTripInput{
					Name:        "Test Trip",
					Description: "Test Description",
					StartDate:   time.Now().Add(24 * time.Hour),
					EndDate:     time.Now().Add(7 * 24 * time.Hour),
					Location:    "Test City",
				},
				setupMocks: func(mockRepo *MockTripRepository) {
					mockRepo.CreateTripFn = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
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
				checkTrip: func(t *testing.T, trip *models.Trip) {
					if trip == nil {
						t.Fatal("Expected trip to be returned")
					}
					if trip.Name != "Test Trip" {
						t.Errorf("Expected name 'Test Trip', got '%s'", trip.Name)
					}
				},
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
				setupMocks: func(mockRepo *MockTripRepository) {
					mockRepo.CreateTripFn = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
						t.Error("Repository should not be called for invalid date range")
						return nil, nil
					}
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
				setupMocks: func(mockRepo *MockTripRepository) {
					mockRepo.CreateTripFn = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
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
				checkTrip: func(t *testing.T, trip *models.Trip) {
					if trip.Name != "Trip to Paris" {
						t.Errorf("Expected name 'Trip to Paris', got '%s'", trip.Name)
					}
				},
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
				setupMocks: func(mockRepo *MockTripRepository) {
					mockRepo.CreateTripFn = func(ctx context.Context, uid uuid.UUID, inp models.CreateTripInput) (*models.Trip, error) {
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
				service, mockTripRepo, _ := setupTestService()

				// Apply mock setup
				tc.setupMocks(mockTripRepo)

				// Execute
				ctx := context.Background()
				userID := uuid.New()
				trip, err := service.CreateTrip(ctx, userID, tc.input)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
					if trip != nil {
						t.Errorf("Expected nil trip, got: %v", trip)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if trip == nil {
						t.Fatal("Expected trip to be returned, got nil")
					}
					if tc.checkTrip != nil {
						tc.checkTrip(t, trip)
					}
				}
			})
		}
	})

	t.Run("GetTripByID", func(t *testing.T) {
		// Table-driven test for GetTripByID
		testCases := []struct {
			name          string
			setupMocks    func(*MockTripRepository, uuid.UUID, uuid.UUID)
			expectedError bool
			errorMessage  string
		}{
			{
				name: "TripNotFound",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return nil, errors.New("trip not found")
					}
				},
				expectedError: true,
				errorMessage:  "trip not found",
			},
			{
				name: "UnauthorizedAccess",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				name: "SuccessfulRetrieval",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return &models.Trip{
							ID:     tripID,
							UserID: userID,
							Name:   "Test Trip",
						}, nil
					}
				},
				expectedError: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				service, mockTripRepo, _ := setupTestService()
				ctx := context.Background()
				tripID := uuid.New()
				userID := uuid.New()

				// Apply mock setup
				tc.setupMocks(mockTripRepo, tripID, userID)

				// Execute
				trip, err := service.GetTripByID(ctx, tripID, userID)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
					if trip != nil {
						t.Errorf("Expected nil trip, got: %v", trip)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if trip == nil {
						t.Fatal("Expected trip to be returned, got nil")
					}
					if trip.ID != tripID || trip.UserID != userID {
						t.Errorf("Trip details don't match expected values")
					}
				}
			})
		}
	})

	t.Run("UpdateTrip", func(t *testing.T) {
		// Table-driven test for UpdateTrip
		now := time.Now()
		testCases := []struct {
			name          string
			updateInput   models.UpdateTripInput
			setupMocks    func(*MockTripRepository, uuid.UUID, uuid.UUID)
			expectedError bool
			errorMessage  string
		}{
			{
				name: "TripNotFound",
				updateInput: models.UpdateTripInput{
					Name: stringPtr("Updated Trip"),
				},
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				name: "InvalidDateRange_BothDates",
				updateInput: models.UpdateTripInput{
					StartDate: timePtr(now.Add(48 * time.Hour)),
					EndDate:   timePtr(now.Add(24 * time.Hour)),
				},
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				name: "ValidUpdate",
				updateInput: models.UpdateTripInput{
					Name:        stringPtr("Updated Trip"),
					Description: stringPtr("Updated Description"),
					StartDate:   timePtr(now.Add(24 * time.Hour)),
					EndDate:     timePtr(now.Add(96 * time.Hour)),
					Location:    stringPtr("Updated City"),
				},
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return &models.Trip{
							ID:        tripID,
							UserID:    userID,
							Name:      "Original Trip",
							StartDate: now,
							EndDate:   now.Add(72 * time.Hour),
						}, nil
					}
					mockRepo.UpdateTripFn = func(ctx context.Context, id uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
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
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				service, mockTripRepo, _ := setupTestService()
				ctx := context.Background()
				tripID := uuid.New()
				userID := uuid.New()

				// Apply mock setup
				tc.setupMocks(mockTripRepo, tripID, userID)

				// Execute
				trip, err := service.UpdateTrip(ctx, tripID, userID, tc.updateInput)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
					if trip != nil {
						t.Errorf("Expected nil trip, got: %v", trip)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if trip == nil {
						t.Fatal("Expected trip to be returned, got nil")
					}
					if *tc.updateInput.Name != trip.Name {
						t.Errorf("Expected updated name '%s', got '%s'", *tc.updateInput.Name, trip.Name)
					}
				}
			})
		}
	})

	t.Run("DeleteTrip", func(t *testing.T) {
		// Table-driven test for DeleteTrip
		testCases := []struct {
			name          string
			setupMocks    func(*MockTripRepository, uuid.UUID, uuid.UUID)
			expectedError bool
			errorMessage  string
		}{
			{
				name: "TripNotFound",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return nil, errors.New("trip not found")
					}
				},
				expectedError: true,
				errorMessage:  "trip not found",
			},
			{
				name: "UnauthorizedAccess",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				name: "SuccessfulDelete",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return &models.Trip{
							ID:     tripID,
							UserID: userID,
						}, nil
					}
					mockRepo.DeleteTripFn = func(ctx context.Context, id uuid.UUID) error {
						return nil
					}
				},
				expectedError: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				service, mockTripRepo, _ := setupTestService()
				ctx := context.Background()
				tripID := uuid.New()
				userID := uuid.New()

				// Apply mock setup
				tc.setupMocks(mockTripRepo, tripID, userID)

				// Execute
				err := service.DeleteTrip(ctx, tripID, userID)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
				}
			})
		}
	})

	t.Run("GetTripWithUser", func(t *testing.T) {
		// Table-driven test for GetTripWithUser
		testCases := []struct {
			name          string
			setupMocks    func(*MockTripRepository, uuid.UUID, uuid.UUID)
			expectedError bool
			errorMessage  string
			checkTrip     func(*testing.T, *models.Trip)
		}{
			{
				name: "TripNotFound",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripWithUserFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return nil, errors.New("trip not found")
					}
				},
				expectedError: true,
				errorMessage:  "trip not found",
			},
			{
				name: "UnauthorizedAccess",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripWithUserFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
						return &models.Trip{
							ID:     tripID,
							UserID: uuid.New(), // Different user
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
			{
				name: "SuccessfulRetrieval",
				setupMocks: func(mockRepo *MockTripRepository, tripID, userID uuid.UUID) {
					mockRepo.GetTripWithUserFn = func(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
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
				checkTrip: func(t *testing.T, trip *models.Trip) {
					if trip.User == nil {
						t.Fatal("Expected user data to be included, got nil")
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				service, mockTripRepo, _ := setupTestService()
				ctx := context.Background()
				tripID := uuid.New()
				userID := uuid.New()

				// Apply mock setup
				tc.setupMocks(mockTripRepo, tripID, userID)

				// Execute
				trip, err := service.GetTripWithUser(ctx, tripID, userID)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
					if trip != nil {
						t.Errorf("Expected nil trip, got: %v", trip)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if trip == nil {
						t.Fatal("Expected trip to be returned, got nil")
					}
					if tc.checkTrip != nil {
						tc.checkTrip(t, trip)
					}
				}
			})
		}
	})

	t.Run("GetUserWithTrips", func(t *testing.T) {
		// Table-driven test for GetUserWithTrips
		testCases := []struct {
			name          string
			setupMocks    func(*MockUserRepository, *MockTripRepository, uuid.UUID)
			expectedError bool
			errorMessage  string
			checkUser     func(*testing.T, *models.User)
		}{
			{
				name: "UserNotFound",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					mockUserRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
						return nil, errors.New("user not found")
					}
				},
				expectedError: true,
				errorMessage:  "user not found",
			},
			{
				name: "SuccessfulRetrieval",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					mockUserRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
						return &models.User{
							ID:   userID,
							Name: "Test User",
						}, nil
					}
					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
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
				checkUser: func(t *testing.T, user *models.User) {
					if len(user.Trips) != 2 {
						t.Errorf("Expected 2 trips, got %d", len(user.Trips))
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				service, mockTripRepo, mockUserRepo := setupTestService()
				ctx := context.Background()
				userID := uuid.New()

				// Apply mock setup
				tc.setupMocks(mockUserRepo, mockTripRepo, userID)

				// Execute
				user, err := service.GetUserWithTrips(ctx, userID, 10, 0)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
					if user != nil {
						t.Errorf("Expected nil user, got: %v", user)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if user == nil {
						t.Fatal("Expected user to be returned, got nil")
					}
					if tc.checkUser != nil {
						tc.checkUser(t, user)
					}
				}
			})
		}
	})

	// New test for GetTripsByUserID
	t.Run("GetTripsByUserID", func(t *testing.T) {
		// Table-driven test for GetTripsByUserID
		testCases := []struct {
			name          string
			setupMocks    func(*MockUserRepository, *MockTripRepository, uuid.UUID)
			expectedError bool
			errorMessage  string
			checkTrips    func(*testing.T, []*models.Trip)
		}{
			{
				name: "UserNotFound",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					mockUserRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
						return nil, errors.New("user not found")
					}
				},
				expectedError: true,
				errorMessage:  "user not found",
			},
			{
				name: "RepositoryError",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					mockUserRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
						return &models.User{
							ID:   userID,
							Name: "Test User",
						}, nil
					}
					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
						return nil, errors.New("database error")
					}
				},
				expectedError: true,
				errorMessage:  "database error",
			},
			{
				name: "EmptyTripList",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					mockUserRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
						return &models.User{
							ID:   userID,
							Name: "Test User",
						}, nil
					}
					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
						return []*models.Trip{}, nil
					}
				},
				expectedError: false,
				checkTrips: func(t *testing.T, trips []*models.Trip) {
					if len(trips) != 0 {
						t.Errorf("Expected empty trip list, got %d trips", len(trips))
					}
				},
			},
			{
				name: "SuccessfulRetrieval",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					mockUserRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
						return &models.User{
							ID:   userID,
							Name: "Test User",
						}, nil
					}
					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, id uuid.UUID, limit, offset int) ([]*models.Trip, error) {
						if limit != 10 || offset != 5 {
							t.Errorf("Expected limit 10 and offset 5, got limit %d and offset %d", limit, offset)
						}
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
							{
								ID:     uuid.New(),
								UserID: userID,
								Name:   "Trip 3",
							},
						}, nil
					}
				},
				expectedError: false,
				checkTrips: func(t *testing.T, trips []*models.Trip) {
					if len(trips) != 3 {
						t.Errorf("Expected 3 trips, got %d", len(trips))
					}

					tripNames := []string{"Trip 1", "Trip 2", "Trip 3"}
					for i, trip := range trips {
						if trip.Name != tripNames[i] {
							t.Errorf("Expected trip name '%s', got '%s'", tripNames[i], trip.Name)
						}
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				service, mockTripRepo, mockUserRepo := setupTestService()
				ctx := context.Background()
				userID := uuid.New()

				// Apply mock setup
				tc.setupMocks(mockUserRepo, mockTripRepo, userID)

				// Execute - use specific pagination values to test they're passed through
				trips, err := service.GetTripsByUserID(ctx, userID, 10, 5)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMessage != "" && err.Error() != tc.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
					}
					if trips != nil {
						t.Errorf("Expected nil trips, got: %v", trips)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if trips == nil {
						t.Fatal("Expected trips to be returned, got nil")
					}
					if tc.checkTrips != nil {
						tc.checkTrips(t, trips)
					}
				}
			})
		}
	})
}
