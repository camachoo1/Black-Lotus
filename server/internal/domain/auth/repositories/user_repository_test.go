package repositories_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	authRepository "black-lotus/internal/domain/auth/repositories"
	tripRepository "black-lotus/internal/domain/trip/repositories"
	"black-lotus/internal/models"
	"black-lotus/pkg/db"
)

// Helper function to check if an error is pgx.ErrNoRows
func isPgxNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

func TestUserRepository(t *testing.T) {
	// Run setup and teardown for each test group
	t.Run("BasicCRUD", func(t *testing.T) {
		setupTestDB(t)
		defer teardownTestDB(t)

		repo := authRepository.NewUserRepository(db.TestDB)

		t.Run("UserCreation", func(t *testing.T) {
			// Table-driven test for user creation scenarios
			testCases := []struct {
				name           string
				input          models.CreateUserInput
				hashedPassword string
				setupFunc      func() // For any pre-test setup
				expectedError  bool
				checkFunc      func(*testing.T, *models.User) // For additional checks
			}{
				{
					name: "ValidUserCreation",
					input: models.CreateUserInput{
						Name:     "Test User",
						Email:    "test@example.com",
						Password: stringPtr("Password123!"),
					},
					hashedPassword: "hashed_password",
					expectedError:  false,
					checkFunc: func(t *testing.T, user *models.User) {
						if user.Name != "Test User" {
							t.Errorf("Expected name %s, got %s", "Test User", user.Name)
						}
						if user.Email != "test@example.com" {
							t.Errorf("Expected email %s, got %s", "test@example.com", user.Email)
						}
						if *user.HashedPassword != "hashed_password" {
							t.Errorf("Expected hashed password %s, got %s", "hashed_password", *user.HashedPassword)
						}
						if user.EmailVerified {
							t.Error("Expected email_verified to be false")
						}
						if user.ID == uuid.Nil {
							t.Error("Expected ID to be set")
						}
						if user.CreatedAt.IsZero() {
							t.Error("Expected CreatedAt to be set")
						}
						if user.UpdatedAt.IsZero() {
							t.Error("Expected UpdatedAt to be set")
						}
					},
				},
				{
					name: "DuplicateEmail",
					input: models.CreateUserInput{
						Name:     "Second User",
						Email:    "duplicate@example.com",
						Password: stringPtr("DifferentPass123!"),
					},
					hashedPassword: "hashed_password",
					setupFunc: func() {
						// Create first user with the same email
						input := models.CreateUserInput{
							Name:     "First User",
							Email:    "duplicate@example.com",
							Password: stringPtr("Password123!"),
						}
						hashedPassword := "hashed_password"
						_, err := repo.CreateUser(context.Background(), input, &hashedPassword)
						if err != nil {
							t.Fatalf("Failed to create first user: %v", err)
						}
					},
					expectedError: true,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Run setup if provided
					if tc.setupFunc != nil {
						tc.setupFunc()
					}

					// Execute test
					user, err := repo.CreateUser(context.Background(), tc.input, &tc.hashedPassword)

					// Verify results
					if tc.expectedError {
						if err == nil {
							t.Error("Expected error, got nil")
						}
						if user != nil {
							t.Errorf("Expected nil user, got: %v", user)
						}
					} else {
						if err != nil {
							t.Fatalf("Expected no error, got: %v", err)
						}
						if user == nil {
							t.Fatal("Expected user to be returned, got nil")
						}
						// Run additional checks
						if tc.checkFunc != nil {
							tc.checkFunc(t, user)
						}
					}
				})
			}
		})

		t.Run("UserLogin", func(t *testing.T) {
			// Table-driven test for login scenarios
			testCases := []struct {
				name          string
				setupFunc     func() // For creating test user
				loginInput    models.LoginUserInput
				expectedError bool
				errorContains string
				checkFunc     func(*testing.T, *models.User) // For additional checks
			}{
				{
					name: "UserNotFound",
					loginInput: models.LoginUserInput{
						Email:    "nonexistent@example.com",
						Password: "Password123!",
					},
					expectedError: true,
					errorContains: "invalid email or password",
				},
				{
					name: "InvalidPassword",
					setupFunc: func() {
						// Create a user with a bcrypt-hashed password
						input := models.CreateUserInput{
							Name:     "Password Test User",
							Email:    "passwordtest@example.com",
							Password: stringPtr("Password123!"),
						}

						// Generate a real bcrypt hash
						hashedBytes, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
						if err != nil {
							t.Fatalf("Failed to hash password: %v", err)
						}
						hashedPassword := string(hashedBytes)

						_, err = repo.CreateUser(context.Background(), input, &hashedPassword)
						if err != nil {
							t.Fatalf("Failed to create test user: %v", err)
						}
					},
					loginInput: models.LoginUserInput{
						Email:    "passwordtest@example.com",
						Password: "WrongPassword!",
					},
					expectedError: true,
					errorContains: "invalid email or password",
				},
				{
					name: "ValidLogin",
					setupFunc: func() {
						// Create a user with bcrypt-hashed password
						input := models.CreateUserInput{
							Name:     "Valid Login User",
							Email:    "validlogin@example.com",
							Password: stringPtr("Password123!"),
						}

						// Generate a bcrypt hash
						password := "Password123!"
						hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
						if err != nil {
							t.Fatalf("Failed to hash password: %v", err)
						}
						hashedPassword := string(hashedBytes)

						_, err = repo.CreateUser(context.Background(), input, &hashedPassword)
						if err != nil {
							t.Fatalf("Failed to create test user: %v", err)
						}
					},
					loginInput: models.LoginUserInput{
						Email:    "validlogin@example.com",
						Password: "Password123!",
					},
					expectedError: false,
					checkFunc: func(t *testing.T, user *models.User) {
						if user.Email != "validlogin@example.com" {
							t.Errorf("Expected email %s, got %s", "validlogin@example.com", user.Email)
						}
						if user.Name != "Valid Login User" {
							t.Errorf("Expected name %s, got %s", "Valid Login User", user.Name)
						}
					},
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Run setup if provided
					if tc.setupFunc != nil {
						tc.setupFunc()
					}

					// Execute test
					user, err := repo.LoginUser(context.Background(), tc.loginInput)

					// Verify results
					if tc.expectedError {
						if err == nil {
							t.Error("Expected error, got nil")
						}
						if user != nil {
							t.Errorf("Expected nil user, got: %v", user)
						}
						if tc.errorContains != "" && (err == nil || !strings.Contains(err.Error(), tc.errorContains)) {
							t.Errorf("Expected error to contain '%s', got: %v", tc.errorContains, err)
						}
					} else {
						if err != nil {
							t.Fatalf("Expected no error, got: %v", err)
						}
						if user == nil {
							t.Fatal("Expected user to be returned, got nil")
						}
						// Run additional checks
						if tc.checkFunc != nil {
							tc.checkFunc(t, user)
						}
					}
				})
			}
		})

		t.Run("GetUserByID", func(t *testing.T) {
			// Create a test user first
			input := models.CreateUserInput{
				Name:     "Get User Test",
				Email:    "getuser@example.com",
				Password: stringPtr("Password123!"),
			}
			hashedPassword := "hashed_password"
			createdUser, err := repo.CreateUser(context.Background(), input, &hashedPassword)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Table-driven test for GetUserByID scenarios
			testCases := []struct {
				name          string
				userID        uuid.UUID
				expectedError bool
				checkFunc     func(*testing.T, *models.User) // For additional checks
			}{
				{
					name:          "ExistingUser",
					userID:        createdUser.ID,
					expectedError: false,
					checkFunc: func(t *testing.T, user *models.User) {
						if user.ID != createdUser.ID {
							t.Errorf("Expected ID %s, got %s", createdUser.ID, user.ID)
						}
						if user.Name != input.Name {
							t.Errorf("Expected name %s, got %s", input.Name, user.Name)
						}
						if user.Email != input.Email {
							t.Errorf("Expected email %s, got %s", input.Email, user.Email)
						}
					},
				},
				{
					name:          "NonExistentUser",
					userID:        uuid.New(),
					expectedError: true,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Execute test
					user, err := repo.GetUserByID(context.Background(), tc.userID)

					// Verify results
					if tc.expectedError {
						if err == nil {
							t.Error("Expected error, got nil")
						}
						if user != nil {
							t.Errorf("Expected nil user, got: %v", user)
						}
						if !isPgxNoRows(err) {
							t.Errorf("Expected pgx.ErrNoRows, got: %v", err)
						}
					} else {
						if err != nil {
							t.Fatalf("Expected no error, got: %v", err)
						}
						if user == nil {
							t.Fatal("Expected user to be returned, got nil")
						}
						// Run additional checks
						if tc.checkFunc != nil {
							tc.checkFunc(t, user)
						}
					}
				})
			}
		})

		t.Run("GetUserByEmail", func(t *testing.T) {
			// Create a test user first
			input := models.CreateUserInput{
				Name:     "Email Test User",
				Email:    "emailtest@example.com",
				Password: stringPtr("Password123!"),
			}
			hashedPassword := "hashed_password"
			_, err := repo.CreateUser(context.Background(), input, &hashedPassword)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Table-driven test for GetUserByEmail scenarios
			testCases := []struct {
				name          string
				email         string
				expectedFound bool
				checkFunc     func(*testing.T, *models.User) // For additional checks
			}{
				{
					name:          "ExistingEmail",
					email:         "emailtest@example.com",
					expectedFound: true,
					checkFunc: func(t *testing.T, user *models.User) {
						if user.Email != input.Email {
							t.Errorf("Expected email %s, got %s", input.Email, user.Email)
						}
						if user.Name != input.Name {
							t.Errorf("Expected name %s, got %s", input.Name, user.Name)
						}
					},
				},
				{
					name:          "NonExistentEmail",
					email:         "nonexistent@example.com",
					expectedFound: false,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Execute test
					user, err := repo.GetUserByEmail(context.Background(), tc.email)

					// Verify results
					if err != nil {
						t.Errorf("Expected nil error for email lookup, got: %v", err)
					}

					if tc.expectedFound {
						if user == nil {
							t.Fatal("Expected user to be returned, got nil")
						}
						// Run additional checks
						if tc.checkFunc != nil {
							tc.checkFunc(t, user)
						}
					} else {
						if user != nil {
							t.Errorf("Expected nil user for non-existent email, got: %v", user)
						}
					}
				})
			}
		})
	})

	t.Run("UserWithTrips", func(t *testing.T) {
		setupTestDB(t)
		defer teardownTestDB(t)

		// Setup repositories
		repo := authRepository.NewUserRepository(db.TestDB)
		tripRepo := tripRepository.NewTripRepository(db.TestDB)

		// Create a test user
		userInput := models.CreateUserInput{
			Name:     "User With Trips",
			Email:    "userwithtrips@example.com",
			Password: stringPtr("Password123!"),
		}
		hashedPassword := "hashed_password"
		user, err := repo.CreateUser(context.Background(), userInput, &hashedPassword)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Create test trips
		for i := 0; i < 5; i++ {
			tripInput := models.CreateTripInput{
				Name:        fmt.Sprintf("Trip %d", i+1),
				Description: fmt.Sprintf("Description for trip %d", i+1),
				StartDate:   time.Now().Add(time.Duration(i*24) * time.Hour),
				EndDate:     time.Now().Add(time.Duration((i+7)*24) * time.Hour),
				Location:    fmt.Sprintf("Location %d", i+1),
			}

			_, err := tripRepo.CreateTrip(context.Background(), user.ID, tripInput)
			if err != nil {
				t.Fatalf("Failed to create test trip: %v", err)
			}
		}

		// Create a second user with no trips for testing
		emptyUserInput := models.CreateUserInput{
			Name:     "User With No Trips",
			Email:    "emptyuser@example.com",
			Password: stringPtr("Password123!"),
		}
		emptyUser, err := repo.CreateUser(context.Background(), emptyUserInput, &hashedPassword)
		if err != nil {
			t.Fatalf("Failed to create empty user: %v", err)
		}

		// Table-driven test for GetUserWithTrips scenarios
		testCases := []struct {
			name          string
			userID        uuid.UUID
			limit         int
			offset        int
			expectedError bool
			expectedTrips int
			checkFunc     func(*testing.T, *models.User) // For additional checks
		}{
			{
				name:          "GetAllTrips",
				userID:        user.ID,
				limit:         10,
				offset:        0,
				expectedError: false,
				expectedTrips: 5,
				checkFunc: func(t *testing.T, user *models.User) {
					// Check that all trips are populated
					tripNames := make(map[string]bool)
					for i := 1; i <= 5; i++ {
						tripNames[fmt.Sprintf("Trip %d", i)] = false
					}

					// Verify each trip belongs to the user and mark the name as found
					for _, trip := range user.Trips {
						if trip.UserID != user.ID {
							t.Errorf("Trip %s: expected user ID %s, got %s", trip.Name, user.ID, trip.UserID)
						}

						// Mark this trip name as found
						if _, exists := tripNames[trip.Name]; exists {
							tripNames[trip.Name] = true
						} else {
							t.Errorf("Unexpected trip name: %s", trip.Name)
						}
					}

					// Verify all expected trip names were found
					for name, found := range tripNames {
						if !found {
							t.Errorf("Expected trip '%s' was not found", name)
						}
					}
				},
			},
			{
				name:          "TestPaginationFirstPage",
				userID:        user.ID,
				limit:         2,
				offset:        0,
				expectedError: false,
				expectedTrips: 2,
			},
			{
				name:          "TestPaginationSecondPage",
				userID:        user.ID,
				limit:         2,
				offset:        2,
				expectedError: false,
				expectedTrips: 2,
			},
			{
				name:          "TestPaginationThirdPage",
				userID:        user.ID,
				limit:         2,
				offset:        4,
				expectedError: false,
				expectedTrips: 1,
			},
			{
				name:          "NonExistentUser",
				userID:        uuid.New(),
				limit:         10,
				offset:        0,
				expectedError: true,
				expectedTrips: 0,
			},
			{
				name:          "UserWithNoTrips",
				userID:        emptyUser.ID,
				limit:         10,
				offset:        0,
				expectedError: false,
				expectedTrips: 0,
			},
		}

		// Store all found trip IDs across pagination tests for uniqueness check
		allTripIDs := make(map[uuid.UUID]bool)

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute test
				userWithTrips, err := repo.GetUserWithTrips(context.Background(), tc.userID, tc.limit, tc.offset)

				// Verify results
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if userWithTrips != nil {
						t.Errorf("Expected nil user, got: %v", userWithTrips)
					}
				} else {
					if err != nil {
						t.Fatalf("Expected no error, got: %v", err)
					}
					if userWithTrips == nil {
						t.Fatal("Expected user to be returned, got nil")
					}
					if userWithTrips.ID != tc.userID {
						t.Errorf("Expected user ID %s, got %s", tc.userID, userWithTrips.ID)
					}
					if userWithTrips.Trips == nil {
						t.Fatal("Expected trips array, got nil")
					}
					if len(userWithTrips.Trips) != tc.expectedTrips {
						t.Errorf("Expected %d trips, got %d", tc.expectedTrips, len(userWithTrips.Trips))
					}

					// Run additional checks
					if tc.checkFunc != nil {
						tc.checkFunc(t, userWithTrips)
					}

					// Record trip IDs for uniqueness check across pagination tests
					if strings.Contains(tc.name, "TestPagination") {
						for _, trip := range userWithTrips.Trips {
							if allTripIDs[trip.ID] {
								t.Errorf("Trip ID %s appears in multiple pages", trip.ID)
							}
							allTripIDs[trip.ID] = true
						}
					}
				}
			})
		}

		// Verify that pagination returned all trips
		if len(allTripIDs) != 5 {
			t.Errorf("Expected 5 unique trip IDs across all pages, got %d", len(allTripIDs))
		}
	})
}
