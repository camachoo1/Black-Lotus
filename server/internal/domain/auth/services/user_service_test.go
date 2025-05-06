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
	users                map[string]*models.User    // map of email to user
	usersByID            map[uuid.UUID]*models.User // map of ID to user
	createUserFunc       func(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error)
	loginUserFunc        func(ctx context.Context, input models.LoginUserInput) (*models.User, error)
	setEmailVerifiedFunc func(ctx context.Context, userID uuid.UUID, verified bool) error
	getUserByIDFunc      func(ctx context.Context, userID uuid.UUID) (*models.User, error)
	getUserByEmailFunc   func(ctx context.Context, email string) (*models.User, error)
	getUserWithTripsFunc func(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error)
}

// NewMockUserRepository creates a new mock repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[string]*models.User),
		usersByID: make(map[uuid.UUID]*models.User),
	}
}

// Implementation of repository methods
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

func (m *MockUserRepository) SetEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	if m.setEmailVerifiedFunc != nil {
		return m.setEmailVerifiedFunc(ctx, userID, verified)
	}

	user, exists := m.usersByID[userID]
	if !exists {
		return errors.New("user not found")
	}

	if !user.EmailVerified {
		user.EmailVerified = verified
	} else {
		return errors.New("user already verified")
	}
	return nil
}

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

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}

	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}

	return user, nil
}

func (m *MockUserRepository) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error) {
	if m.getUserWithTripsFunc != nil {
		return m.getUserWithTripsFunc(ctx, userID, limit, offset)
	}

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

// Implementation of trip repository methods
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

// Implement other required methods to satisfy the interface
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

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// setupTestUser creates a test user with password in the mock repositories
func setupTestUser() (*MockUserRepository, *MockTripRepository, *models.User, string) {
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

	return mockUserRepo, mockTripRepo, user, password
}

func TestUserService(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		// Table-driven test for user creation scenarios
		testCases := []struct {
			name          string
			input         models.CreateUserInput
			setupMocks    func(*MockUserRepository)
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
				setupMocks:    func(*MockUserRepository) {},
				expectedError: false,
			},
			{
				name: "DuplicateEmail",
				input: models.CreateUserInput{
					Name:     "Second User",
					Email:    "duplicate@example.com",
					Password: stringPtr("DifferentPass123!"),
				},
				setupMocks: func(repo *MockUserRepository) {
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
				setupMocks: func(repo *MockUserRepository) {
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
				mockUserRepo := NewMockUserRepository()
				mockTripRepo := NewMockTripRepository()

				// Apply any custom mock setup
				tc.setupMocks(mockUserRepo)

				service := services.NewUserService(mockUserRepo, mockTripRepo)

				// Execute
				user, err := service.CreateUser(context.Background(), tc.input)

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
	})

	t.Run("LoginUser", func(t *testing.T) {
		// Table-driven test for login scenarios
		testCases := []struct {
			name          string
			inputEmail    string
			inputPassword string
			setupMocks    func() (*MockUserRepository, *MockTripRepository, *models.User)
			expectedError bool
			errorMessage  string
		}{
			{
				name:          "SuccessfulLogin",
				inputEmail:    "test@example.com",
				inputPassword: "Password123!",
				setupMocks: func() (*MockUserRepository, *MockTripRepository, *models.User) {
					mockUserRepo, mockTripRepo, testUser, _ := setupTestUser()
					return mockUserRepo, mockTripRepo, testUser
				},
				expectedError: false,
			},
			{
				name:          "InvalidPassword",
				inputEmail:    "test@example.com",
				inputPassword: "WrongPassword!",
				setupMocks: func() (*MockUserRepository, *MockTripRepository, *models.User) {
					mockUserRepo, mockTripRepo, testUser, _ := setupTestUser()
					return mockUserRepo, mockTripRepo, testUser
				},
				expectedError: true,
				errorMessage:  "invalid email or password",
			},
			{
				name:          "NonExistentUser",
				inputEmail:    "nonexistent@example.com",
				inputPassword: "Password123!",
				setupMocks: func() (*MockUserRepository, *MockTripRepository, *models.User) {
					mockUserRepo, mockTripRepo, testUser, _ := setupTestUser()
					return mockUserRepo, mockTripRepo, testUser
				},
				expectedError: true,
				errorMessage:  "invalid email or password",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				mockUserRepo, mockTripRepo, expectedUser := tc.setupMocks()
				service := services.NewUserService(mockUserRepo, mockTripRepo)

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
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Table-driven test for GetUserByID scenarios
		testCases := []struct {
			name          string
			userExists    bool
			expectedError bool
		}{
			{
				name:          "ExistingUser",
				userExists:    true,
				expectedError: false,
			},
			{
				name:          "NonExistentUser",
				userExists:    false,
				expectedError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				mockUserRepo := NewMockUserRepository()
				mockTripRepo := NewMockTripRepository()

				var userID uuid.UUID
				if tc.userExists {
					userID = uuid.New()
					hashedPassword := "hashed_password"
					mockUserRepo.usersByID[userID] = &models.User{
						ID:             userID,
						Name:           "Test User",
						Email:          "test@example.com",
						HashedPassword: &hashedPassword,
						EmailVerified:  false,
					}
				} else {
					userID = uuid.New() // Random non-existent ID
				}

				service := services.NewUserService(mockUserRepo, mockTripRepo)

				// Execute
				user, err := service.GetUserByID(context.Background(), userID)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
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
					if user.ID != userID {
						t.Errorf("Expected user ID %s, got %s", userID, user.ID)
					}
					if user.HashedPassword != nil {
						t.Error("Expected hashed password to be nil in returned user")
					}
				}
			})
		}
	})

	t.Run("GetUserWithTrips", func(t *testing.T) {
		// Table-driven test for different trip scenarios
		testCases := []struct {
			name          string
			setupMocks    func(*MockUserRepository, *MockTripRepository, uuid.UUID)
			expectedError bool
			tripCount     int
		}{
			{
				name: "UserWithNoTrips",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					hashedPassword := "hashed_password"
					mockUserRepo.usersByID[userID] = &models.User{
						ID:             userID,
						Name:           "Test User",
						Email:          "test@example.com",
						HashedPassword: &hashedPassword,
						EmailVerified:  false,
					}

					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
						return []*models.Trip{}, nil
					}
				},
				expectedError: false,
				tripCount:     0,
			},
			{
				name: "UserWithTrips",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					hashedPassword := "hashed_password"
					mockUserRepo.usersByID[userID] = &models.User{
						ID:             userID,
						Name:           "Test User",
						Email:          "test@example.com",
						HashedPassword: &hashedPassword,
						EmailVerified:  false,
					}

					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
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
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					// Don't add any user - simulating non-existent user
				},
				expectedError: true,
				tripCount:     0,
			},
			{
				name: "ErrorGettingTrips",
				setupMocks: func(mockUserRepo *MockUserRepository, mockTripRepo *MockTripRepository, userID uuid.UUID) {
					hashedPassword := "hashed_password"
					mockUserRepo.usersByID[userID] = &models.User{
						ID:             userID,
						Name:           "Test User",
						Email:          "test@example.com",
						HashedPassword: &hashedPassword,
						EmailVerified:  false,
					}

					mockTripRepo.GetTripsByUserIDFn = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
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
				mockUserRepo := NewMockUserRepository()
				mockTripRepo := NewMockTripRepository()
				userID := uuid.New()

				// Apply custom mock setup
				tc.setupMocks(mockUserRepo, mockTripRepo, userID)

				service := services.NewUserService(mockUserRepo, mockTripRepo)

				// Execute
				user, err := service.GetUserWithTrips(context.Background(), userID, 10, 0)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
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
	})
}
