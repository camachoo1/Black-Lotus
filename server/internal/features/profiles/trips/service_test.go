// internal/features/profiles/trips/service_test.go
package trips_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/profiles/trips"
)

// MockUserRepository implements trips.UserRepository for testing
type MockUserRepository struct {
	getUserByIDFunc func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented")
}

// MockTripRepository implements trips.TripRepository for testing
type MockTripRepository struct {
	getTripsByUserIDFunc func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
}

func (m *MockTripRepository) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error) {
	if m.getTripsByUserIDFunc != nil {
		return m.getTripsByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetTripsByUserID not implemented")
}

func TestGetUserWithTrips(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*testing.T, *MockUserRepository, *MockTripRepository, uuid.UUID)
		expectedError bool
		tripCount     int
	}{
		{
			name: "UserWithNoTrips",
			setupMocks: func(t *testing.T, mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
				hashedPassword := "hashed_password"
				mockUserRepo.getUserByIDFunc = func(ctx context.Context, uid uuid.UUID) (*models.User, error) {
					if uid == userID {
						return &models.User{
							ID:             userID,
							Name:           "Test User",
							Email:          "test@example.com",
							HashedPassword: &hashedPassword,
							EmailVerified:  false,
						}, nil
					}
					return nil, errors.New("user not found")
				}

				mockTripRepo.getTripsByUserIDFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return []*models.Trip{}, nil
				}
			},
			expectedError: false,
			tripCount:     0,
		},
		{
			name: "UserWithTrips",
			setupMocks: func(t *testing.T, mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
				hashedPassword := "hashed_password"
				mockUserRepo.getUserByIDFunc = func(ctx context.Context, uid uuid.UUID) (*models.User, error) {
					if uid == userID {
						return &models.User{
							ID:             userID,
							Name:           "Test User",
							Email:          "test@example.com",
							HashedPassword: &hashedPassword,
							EmailVerified:  false,
						}, nil
					}
					return nil, errors.New("user not found")
				}

				mockTripRepo.getTripsByUserIDFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return []*models.Trip{
						{
							ID:          uuid.New(),
							UserID:      userID,
							Name:        "Trip to Paris",
							Description: "Vacation in Paris",
							Location:    "Paris",
						},
						{
							ID:          uuid.New(),
							UserID:      userID,
							Name:        "Trip to Rome",
							Description: "Business trip to Rome",
							Location:    "Rome",
						},
					}, nil
				}
			},
			expectedError: false,
			tripCount:     2,
		},
		{
			name: "NonExistentUser",
			setupMocks: func(t *testing.T, mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
				mockUserRepo.getUserByIDFunc = func(ctx context.Context, uid uuid.UUID) (*models.User, error) {
					return nil, errors.New("user not found")
				}
			},
			expectedError: true,
			tripCount:     0,
		},
		{
			name: "NilUserReturned",
			setupMocks: func(t *testing.T, mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
				mockUserRepo.getUserByIDFunc = func(ctx context.Context, uid uuid.UUID) (*models.User, error) {
					return nil, nil
				}
			},
			expectedError: true,
			tripCount:     0,
		},
		{
			name: "ErrorGettingTrips",
			setupMocks: func(t *testing.T, mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
				hashedPassword := "hashed_password"
				mockUserRepo.getUserByIDFunc = func(ctx context.Context, uid uuid.UUID) (*models.User, error) {
					if uid == userID {
						return &models.User{
							ID:             userID,
							Name:           "Test User",
							Email:          "test@example.com",
							HashedPassword: &hashedPassword,
							EmailVerified:  false,
						}, nil
					}
					return nil, errors.New("user not found")
				}

				mockTripRepo.getTripsByUserIDFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: true,
			tripCount:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockUserRepo := &MockUserRepository{}
			mockTripRepo := &MockTripRepository{}
			userID := uuid.New()

			// Apply custom mock setup
			tc.setupMocks(t, mockUserRepo, mockTripRepo, userID)

			service := trips.NewService(mockUserRepo, mockTripRepo)

			// Execute
			user, err := service.GetUserWithTrips(context.Background(), userID, 10, 0)

			// Verify
			if tc.expectedError {
				if err == nil && user == nil {
					// This is fine - nil user without an error is handled as expected
				} else if err == nil {
					t.Error("Expected error or nil user, got user without error")
				}

				if user != nil && err == nil {
					t.Errorf("Expected nil user or error, got: %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if user == nil {
					t.Fatal("Expected user to be returned, got nil")
				}
				if user.ID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, user.ID)
				}
				if user.HashedPassword != nil {
					t.Error("Expected hashed password to be nil in returned user")
				}
				if user.Trips == nil {
					t.Error("Expected trips array, got nil")
				}
				if len(user.Trips) != tc.tripCount {
					t.Errorf("Expected %d trips, got %d", tc.tripCount, len(user.Trips))
				}

				if tc.tripCount > 0 {
					// Check trip names if we expect trips
					tripNames := []string{"Trip to Paris", "Trip to Rome"}
					for i, trip := range user.Trips {
						if trip.Name != tripNames[i] {
							t.Errorf("Expected trip %d to be named %s, got %s", i, tripNames[i], trip.Name)
						}
					}
				}
			}
		})
	}
}
