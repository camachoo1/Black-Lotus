// internal/features/auth/register/service_test.go
package register_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/register"
)

// MockRepository implements register.Repository for testing
type MockRepository struct {
	users          map[string]*models.User
	usersByID      map[uuid.UUID]*models.User
	createUserFunc func(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error)
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:     make(map[string]*models.User),
		usersByID: make(map[uuid.UUID]*models.User),
	}
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockRepository) CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, input, hashedPassword)
	}

	// Check if a user already exists with this email
	if _, exists := m.users[input.Email]; exists {
		return nil, errors.New("user with this email already exists")
	}

	id := uuid.New()
	user := &models.User{
		ID:             id,
		Name:           input.Name,
		Email:          input.Email,
		HashedPassword: hashedPassword,
		EmailVerified:  false,
	}

	m.users[input.Email] = user
	m.usersByID[id] = user

	return user, nil
}

func (m *MockRepository) SetEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	user, exists := m.usersByID[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.EmailVerified = verified
	return nil
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

func TestRegisterService(t *testing.T) {
	testCases := []struct {
		name          string
		input         models.CreateUserInput
		setupMocks    func(*MockRepository)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "SuccessfulUserCreation",
			input: models.CreateUserInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: stringPtr("Password123!"),
			},
			setupMocks:    func(*MockRepository) {},
			expectedError: false,
		},
		{
			name: "DuplicateEmail",
			input: models.CreateUserInput{
				Name:     "Second User",
				Email:    "duplicate@example.com",
				Password: stringPtr("DifferentPass123!"),
			},
			setupMocks: func(repo *MockRepository) {
				// Pre-populate with a user with the same email
				hashedPassword := "hashed_password"
				repo.users["duplicate@example.com"] = &models.User{
					ID:             uuid.New(),
					Name:           "First User",
					Email:          "duplicate@example.com",
					HashedPassword: &hashedPassword,
				}
			},
			expectedError: true,
			errorMessage:  "user with this email already exists",
		},
		{
			name: "RepositoryError",
			input: models.CreateUserInput{
				Name:     "Error User",
				Email:    "error@example.com",
				Password: stringPtr("Password123!"),
			},
			setupMocks: func(repo *MockRepository) {
				repo.createUserFunc = func(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
					return nil, errors.New("repository error")
				}
			},
			expectedError: true,
			errorMessage:  "repository error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockRepo := NewMockRepository()

			// Apply any custom mock setup
			tc.setupMocks(mockRepo)

			service := register.NewService(mockRepo)

			// Execute
			user, err := service.Register(context.Background(), tc.input)

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
				if user.Name != tc.input.Name {
					t.Errorf("Expected name %s, got %s", tc.input.Name, user.Name)
				}
				if user.Email != tc.input.Email {
					t.Errorf("Expected email %s, got %s", tc.input.Email, user.Email)
				}
				if user.HashedPassword != nil {
					t.Error("Expected hashed password to be nil in returned user")
				}
			}
		})
	}
}
