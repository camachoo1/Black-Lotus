package view_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/profiles/view"
)

// MockRepository implements view.Repository for testing
type MockRepository struct {
	getUserByIDFunc func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *MockRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented")
}

// Helper function to setup service for testing
func setupServiceTest() (*view.Service, *MockRepository) {
	mockRepo := &MockRepository{}
	service := view.NewService(mockRepo)
	return service, mockRepo
}

func TestServiceGetUserProfile(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*testing.T, *MockRepository, uuid.UUID) *models.User
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulGetProfile",
			setupMock: func(t *testing.T, repo *MockRepository, userID uuid.UUID) *models.User {
				hashedPassword := "hashed_password"
				expectedUser := &models.User{
					ID:             userID,
					Name:           "Test User",
					Email:          "test@example.com",
					HashedPassword: &hashedPassword,
					EmailVerified:  true,
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
			errorMessage:  "",
		},
		{
			name: "UserNotFound",
			setupMock: func(t *testing.T, repo *MockRepository, userID uuid.UUID) *models.User {
				repo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("user not found")
				}
				return nil
			},
			expectedError: true,
			errorMessage:  "user not found",
		},
		{
			name: "NilUserReturned",
			setupMock: func(t *testing.T, repo *MockRepository, userID uuid.UUID) *models.User {
				repo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, nil
				}
				return nil
			},
			expectedError: false,
			errorMessage:  "",
		},
		{
			name: "DatabaseError",
			setupMock: func(t *testing.T, repo *MockRepository, userID uuid.UUID) *models.User {
				repo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("database error")
				}
				return nil
			},
			expectedError: true,
			errorMessage:  "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo := setupServiceTest()
			userID := uuid.New()

			// Setup mock
			expectedUser := tc.setupMock(t, mockRepo, userID)

			// Execute
			result, err := service.GetUserProfile(context.Background(), userID)

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

				// Special case for NilUserReturned test
				if tc.name == "NilUserReturned" {
					if result != nil {
						t.Error("Expected nil result, got non-nil")
					}
					return
				}

				if result == nil {
					t.Fatal("Expected user to be returned, got nil")
				}

				if expectedUser == nil {
					return // Skip further checks if no expected user
				}

				if result.ID != expectedUser.ID {
					t.Errorf("Expected user ID %s, got %s", expectedUser.ID, result.ID)
				}

				if result.HashedPassword != nil {
					t.Error("Expected hashed password to be nil in returned user")
				}

				if result.Name != expectedUser.Name {
					t.Errorf("Expected name %s, got %s", expectedUser.Name, result.Name)
				}

				if result.Email != expectedUser.Email {
					t.Errorf("Expected email %s, got %s", expectedUser.Email, result.Email)
				}
			}
		})
	}
}
