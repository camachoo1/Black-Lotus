package login_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/login"
)

// MockRepository implements login.Repository for testing
type MockRepository struct {
	users              map[string]*models.User
	loginUserFunc      func(ctx context.Context, input models.LoginUserInput) (*models.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockRepository) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	// If a custom implementation is provided, use it
	if m.loginUserFunc != nil {
		return m.loginUserFunc(ctx, input)
	}

	// Otherwise use the default implementation that checks the users map
	user, exists := m.users[input.Email]
	if !exists {
		return nil, errors.New("invalid email or password")
	}

	// Verify password if hashedPassword exists
	if user.HashedPassword != nil {
		err := bcrypt.CompareHashAndPassword([]byte(*user.HashedPassword), []byte(input.Password))
		if err != nil {
			return nil, errors.New("invalid email or password")
		}
	}

	return user, nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// If a custom implementation is provided, use it
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}

	// Otherwise use the default implementation that checks the users map
	user, exists := m.users[email]
	if !exists {
		return nil, nil // User not found, same as the real implementation
	}

	return user, nil
}

// setupTestUser creates a test user with password
func setupTestUser() (*MockRepository, *models.User, string) {
	mockRepo := NewMockRepository()

	// Create a user with a real bcrypt password
	password := "Password123!"
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPassword := string(hashedBytes)

	userID := uuid.New()
	user := &models.User{
		ID:             userID,
		Name:           "Test User",
		Email:          "test@example.com",
		HashedPassword: &hashedPassword,
		EmailVerified:  false,
	}

	mockRepo.users[user.Email] = user

	return mockRepo, user, password
}

func TestLoginService(t *testing.T) {
	testCases := []struct {
		name          string
		inputEmail    string
		inputPassword string
		setupMocks    func() (*MockRepository, *models.User)
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "SuccessfulLogin",
			inputEmail:    "test@example.com",
			inputPassword: "Password123!",
			setupMocks: func() (*MockRepository, *models.User) {
				mockRepo, testUser, _ := setupTestUser()
				// No need to set loginUserFunc, we'll use the default implementation
				return mockRepo, testUser
			},
			expectedError: false,
		},
		{
			name:          "InvalidPassword",
			inputEmail:    "test@example.com",
			inputPassword: "WrongPassword!",
			setupMocks: func() (*MockRepository, *models.User) {
				mockRepo, testUser, _ := setupTestUser()
				// No need to set loginUserFunc, we'll use the default implementation
				return mockRepo, testUser
			},
			expectedError: true,
			errorMessage:  "invalid email or password",
		},
		{
			name:          "NonExistentUser",
			inputEmail:    "nonexistent@example.com",
			inputPassword: "Password123!",
			setupMocks: func() (*MockRepository, *models.User) {
				mockRepo, testUser, _ := setupTestUser()
				// No need to set loginUserFunc, we'll use the default implementation
				return mockRepo, testUser
			},
			expectedError: true,
			errorMessage:  "invalid email or password",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockRepo, expectedUser := tc.setupMocks()
			service := login.NewService(mockRepo)

			// Create login input
			input := models.LoginUserInput{
				Email:    tc.inputEmail,
				Password: tc.inputPassword,
			}

			// Execute
			user, err := service.LoginUser(context.Background(), input)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
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
				if user.ID != expectedUser.ID {
					t.Errorf("Expected user ID %s, got %s", expectedUser.ID, user.ID)
				}
			}
		})
	}
}
