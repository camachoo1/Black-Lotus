package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/user"
)

// MockRepository is defined in handler_test.go

// Helper function to setup service for testing
func setupServiceTest() (user.ServiceInterface, *MockRepository) {
	mockRepo := &MockRepository{}
	service := user.NewService(mockRepo)
	return service, mockRepo
}

func TestServiceGetUserByID(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*testing.T, *MockRepository) *models.User
		expectedError bool
		errorMessage  string
		checkPassword bool
	}{
		{
			name: "SuccessfulGetUser",
			setupMock: func(t *testing.T, repo *MockRepository) *models.User {
				userID := uuid.New()
				hashedPassword := "hashed_password"
				expectedUser := &models.User{
					ID:             userID,
					Name:           "Test User",
					Email:          "test@example.com",
					HashedPassword: &hashedPassword,
					EmailVerified:  false,
				}

				repo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					if id == userID {
						userCopy := *expectedUser // Copy to avoid modifying the original
						return &userCopy, nil
					}
					return nil, errors.New("user not found")
				}

				return expectedUser
			},
			expectedError: false,
			checkPassword: true,
		},
		{
			name: "UserNotFound",
			setupMock: func(t *testing.T, repo *MockRepository) *models.User {
				repo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("user not found")
				}
				return nil
			},
			expectedError: true,
			errorMessage:  "user not found",
		},
		{
			name: "UserWithoutPassword",
			setupMock: func(t *testing.T, repo *MockRepository) *models.User {
				userID := uuid.New()
				expectedUser := &models.User{
					ID:             userID,
					Name:           "Test User",
					Email:          "test@example.com",
					HashedPassword: nil,
					EmailVerified:  false,
				}

				repo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					if id == userID {
						userCopy := *expectedUser
						return &userCopy, nil
					}
					return nil, errors.New("user not found")
				}

				return expectedUser
			},
			expectedError: false,
			checkPassword: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo := setupServiceTest()
			expectedUser := tc.setupMock(t, mockRepo)

			var userID uuid.UUID
			if expectedUser != nil {
				userID = expectedUser.ID
			} else {
				userID = uuid.New() // Random ID for error cases
			}

			// Execute
			result, err := service.GetUserByID(context.Background(), userID)

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
					t.Fatal("Expected user to be returned, got nil")
				}
				if result.ID != expectedUser.ID {
					t.Errorf("Expected user ID %s, got %s", expectedUser.ID, result.ID)
				}
				if result.HashedPassword != nil {
					t.Error("Expected hashed password to be nil in returned user")
				}
			}
		})
	}
}
