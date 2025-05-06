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

// GetUserWithTrips mocks getting a user with their trips
func (m *MockUserRepository) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error) {
	user, exists := m.usersByID[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// For testing purposes, we're not actually populating trips
	// In a real implementation, this would load trips from a repository
	user.Trips = []*models.Trip{}

	return user, nil
}

// Mock TripRepository for testing
type MockTripRepository struct {
	trips              []*models.Trip
	GetTripsByUserIDFn func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
}

// NewMockTripRepository creates a new mock trip repository
func NewMockTripRepository() *MockTripRepository {
	return &MockTripRepository{
		trips: make([]*models.Trip, 0),
	}
}

// GetTripsByUserID mocks getting trips for a user
func (m *MockTripRepository) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error) {
	if m.GetTripsByUserIDFn != nil {
		return m.GetTripsByUserIDFn(ctx, userID, limit, offset)
	}

	userTrips := make([]*models.Trip, 0)
	for _, trip := range m.trips {
		if trip.UserID == userID {
			userTrips = append(userTrips, trip)
		}
	}

	return userTrips, nil
}

// Ensure MockTripRepository fully implements the interface
func (m *MockTripRepository) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	return nil, errors.New("CreateTrip not implemented")
}

func (m *MockTripRepository) GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	return nil, errors.New("GetTripByID not implemented")
}

func (m *MockTripRepository) UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	return nil, errors.New("UpdateTrip not implemented")
}

func (m *MockTripRepository) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	return errors.New("DeleteTrip not implemented")
}

func (m *MockTripRepository) GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	return nil, errors.New("GetTripWithUser not implemented")
}

func TestCreateUser(t *testing.T) {
	t.Run("Create New User", func(t *testing.T) {
		// Setup
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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
		mockUserRepo := NewMockUserRepository()
		mockUserRepo.createUserFunc = func(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
			// Simulate error in repository layer
			return nil, errors.New("repository error")
		}
		mockTripRepo := NewMockTripRepository()
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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

// Test for GetUserWithTrips
func TestGetUserWithTrips(t *testing.T) {
	t.Run("User With No Trips", func(t *testing.T) {
		// Setup
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()

		userID := uuid.New()
		hashedPassword := "hashed_password"
		mockUserRepo.usersByID[userID] = &models.User{
			ID:             userID,
			Name:           "Test User",
			Email:          "test@example.com",
			HashedPassword: &hashedPassword,
			EmailVerified:  false,
		}

		// Configure the mock to return empty trips
		mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
			return []*models.Trip{}, nil
		}

		service := services.NewUserService(mockUserRepo, mockTripRepo)

		// Execute
		user, err := service.GetUserWithTrips(context.Background(), userID, 10, 0)

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

		if user.Trips == nil {
			t.Error("Expected empty trips array, got nil")
		}

		if len(user.Trips) > 0 {
			t.Errorf("Expected 0 trips, got %d", len(user.Trips))
		}
	})

	t.Run("User With Trips", func(t *testing.T) {
		// Setup
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()

		userID := uuid.New()
		hashedPassword := "hashed_password"
		mockUserRepo.usersByID[userID] = &models.User{
			ID:             userID,
			Name:           "Test User",
			Email:          "test@example.com",
			HashedPassword: &hashedPassword,
			EmailVerified:  false,
		}

		// Configure the mock to return trips
		tripID1 := uuid.New()
		tripID2 := uuid.New()
		mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
			return []*models.Trip{
				{
					ID:          tripID1,
					UserID:      userID,
					Name:        "Trip to Paris",
					Description: "Vacation in Paris",
					Location:    "Paris",
				},
				{
					ID:          tripID2,
					UserID:      userID,
					Name:        "Trip to Rome",
					Description: "Business trip to Rome",
					Location:    "Rome",
				},
			}, nil
		}

		service := services.NewUserService(mockUserRepo, mockTripRepo)

		// Execute
		user, err := service.GetUserWithTrips(context.Background(), userID, 10, 0)

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if user.Trips == nil {
			t.Fatal("Expected trips array, got nil")
		}

		if len(user.Trips) != 2 {
			t.Errorf("Expected 2 trips, got %d", len(user.Trips))
		}

		if user.Trips[0].Name != "Trip to Paris" || user.Trips[1].Name != "Trip to Rome" {
			t.Errorf("Trips were not correctly populated")
		}
	})

	t.Run("Non-existent User", func(t *testing.T) {
		// Setup
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()
		service := services.NewUserService(mockUserRepo, mockTripRepo)
		nonExistentID := uuid.New()

		// Execute
		user, err := service.GetUserWithTrips(context.Background(), nonExistentID, 10, 0)

		// Verify
		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}
	})

	t.Run("Error Getting Trips", func(t *testing.T) {
		// Setup
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()

		userID := uuid.New()
		hashedPassword := "hashed_password"
		mockUserRepo.usersByID[userID] = &models.User{
			ID:             userID,
			Name:           "Test User",
			Email:          "test@example.com",
			HashedPassword: &hashedPassword,
			EmailVerified:  false,
		}

		// Configure the mock to return an error
		mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
			return nil, errors.New("database error")
		}

		service := services.NewUserService(mockUserRepo, mockTripRepo)

		// Execute
		user, err := service.GetUserWithTrips(context.Background(), userID, 10, 0)

		// Verify
		if err == nil {
			t.Error("Expected error getting trips, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user when trips error, got: %v", user)
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got: %v", err)
		}
	})
}

func TestLoginUser(t *testing.T) {
	// Setup common mock repository with a test user
	setupTestUser := func() (*MockUserRepository, *MockTripRepository, *models.User) {
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()

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

		mockUserRepo.users[user.Email] = user
		mockUserRepo.usersByID[userID] = user

		return mockUserRepo, mockTripRepo, user
	}

	t.Run("Successful Login", func(t *testing.T) {
		// Setup
		mockUserRepo, mockTripRepo, testUser := setupTestUser()
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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
		mockUserRepo, mockTripRepo, _ := setupTestUser()
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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
		mockUserRepo, mockTripRepo, _ := setupTestUser()
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()
		userID := uuid.New()
		hashedPassword := "hashed_password"
		mockUserRepo.usersByID[userID] = &models.User{
			ID:             userID,
			Name:           "Test User",
			Email:          "test@example.com",
			HashedPassword: &hashedPassword,
			EmailVerified:  false,
		}
		service := services.NewUserService(mockUserRepo, mockTripRepo)

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
		mockUserRepo := NewMockUserRepository()
		mockTripRepo := NewMockTripRepository()
		service := services.NewUserService(mockUserRepo, mockTripRepo)
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
