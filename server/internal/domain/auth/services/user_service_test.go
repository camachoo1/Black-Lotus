package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"black-lotus/internal/domain/auth/services"
	"black-lotus/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users           map[string]*models.User    // map of email to user
	usersByID       map[uuid.UUID]*models.User // map of ID to user
	createUserFunc  func(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error)
	loginUserFunc   func(ctx context.Context, input models.LoginUserInput) (*models.User, error)
	getUserByIDFunc func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

// NewMockUserRepository creates a new mock repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[string]*models.User),
		usersByID: make(map[uuid.UUID]*models.User),
	}
}

// CreateUser mocks the repository's CreateUser method
func (m *MockUserRepository) CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, input, hashedPassword)
	}

	// Default implementation
	for _, existingUser := range m.users {
		if existingUser.Email == input.Email {
			return nil, errors.New("user with this email already exists")
		}
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

// LoginUser mocks the repository's LoginUser method
func (m *MockUserRepository) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	if m.loginUserFunc != nil {
		return m.loginUserFunc(ctx, input)
	}

	// Default implementation
	user, exists := m.users[input.Email]
	if !exists {
		return nil, errors.New("invalid email or password")
	}

	if user.HashedPassword == nil {
		return nil, errors.New("user has no password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(*user.HashedPassword), []byte(input.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

// GetUserByID mocks the repository's GetUserByID method
func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}

	// Default implementation
	user, exists := m.usersByID[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail mocks the repository's GetUserByEmail method
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}

	return user, nil
}

func TestCreateUser(t *testing.T) {
	t.Run("Create New User", func(t *testing.T) {
		// Setup
		mockRepo := NewMockUserRepository()
		service := services.NewUserService(mockRepo)

		// Input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}

		// Execute
		user, err := service.CreateUser(context.Background(), input)

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if user.Name != input.Name {
			t.Errorf("Expected name %s, got %s", input.Name, user.Name)
		}

		if user.Email != input.Email {
			t.Errorf("Expected email %s, got %s", input.Email, user.Email)
		}

		if user.HashedPassword != nil {
			t.Error("Expected hashed password to be nil in returned user")
		}
	})

	t.Run("Create User With Existing Email", func(t *testing.T) {
		// Setup
		mockRepo := NewMockUserRepository()
		service := services.NewUserService(mockRepo)

		// Create first user
		input1 := models.CreateUserInput{
			Name:     "First User",
			Email:    "duplicate@example.com",
			Password: stringPtr("Password123!"),
		}

		_, err := service.CreateUser(context.Background(), input1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create second user with same email
		input2 := models.CreateUserInput{
			Name:     "Second User",
			Email:    "duplicate@example.com", // Same email
			Password: stringPtr("DifferentPass123!"),
		}

		user, err := service.CreateUser(context.Background(), input2)

		// Verify
		if err == nil {
			t.Error("Expected error for duplicate email, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}

		expectedError := "user with this email already exists"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("Create User With Password Hashing Error", func(t *testing.T) {
		// Setup
		mockRepo := NewMockUserRepository()
		mockRepo.createUserFunc = func(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
			// Simulate error in repository layer
			return nil, errors.New("repository error")
		}
		service := services.NewUserService(mockRepo)

		// Input
		input := models.CreateUserInput{
			Name:     "Error User",
			Email:    "error@example.com",
			Password: stringPtr("Password123!"),
		}

		// Execute
		user, err := service.CreateUser(context.Background(), input)

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}
	})
}

func TestLoginUser(t *testing.T) {
	// Setup common mock repository with a test user
	setupTestUser := func() (*MockUserRepository, *models.User) {
		mockRepo := NewMockUserRepository()

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
		mockRepo.usersByID[userID] = user

		return mockRepo, user
	}

	t.Run("Successful Login", func(t *testing.T) {
		// Setup
		mockRepo, testUser := setupTestUser()
		service := services.NewUserService(mockRepo)

		// Input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		// Execute
		user, err := service.LoginUser(context.Background(), input)

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %s, got %s", testUser.ID, user.ID)
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		// Setup
		mockRepo, _ := setupTestUser()
		service := services.NewUserService(mockRepo)

		// Input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "WrongPassword!",
		}

		// Execute
		user, err := service.LoginUser(context.Background(), input)

		// Verify
		if err == nil {
			t.Error("Expected error for wrong password, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}

		expectedError := "invalid email or password"
		if err != nil && err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("Non-existent User", func(t *testing.T) {
		// Setup
		mockRepo, _ := setupTestUser()
		service := services.NewUserService(mockRepo)

		// Input
		input := models.LoginUserInput{
			Email:    "nonexistent@example.com",
			Password: "Password123!",
		}

		// Execute
		user, err := service.LoginUser(context.Background(), input)

		// Verify
		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}

		expectedError := "invalid email or password"
		if err != nil && err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("Existing User", func(t *testing.T) {
		// Setup
		mockRepo := NewMockUserRepository()
		userID := uuid.New()
		hashedPassword := "hashed_password"
		mockRepo.usersByID[userID] = &models.User{
			ID:             userID,
			Name:           "Test User",
			Email:          "test@example.com",
			HashedPassword: &hashedPassword,
			EmailVerified:  false,
		}
		service := services.NewUserService(mockRepo)

		// Execute
		user, err := service.GetUserByID(context.Background(), userID)

		// Verify
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
	})

	t.Run("Non-existent User", func(t *testing.T) {
		// Setup
		mockRepo := NewMockUserRepository()
		service := services.NewUserService(mockRepo)
		nonExistentID := uuid.New()

		// Execute
		user, err := service.GetUserByID(context.Background(), nonExistentID)

		// Verify
		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
