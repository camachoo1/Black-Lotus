package repositories_test

import (
	"context"
	"errors"
	"fmt"
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

// setupTestDB ensures the test database is initialized
func setupTestDB(t *testing.T) {
	// Initialize the test database
	if err := db.InitializeTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Clean up all tables before each test
	if err := db.CleanTestTables(context.Background()); err != nil {
		t.Fatalf("Failed to clean test tables: %v", err)
	}
}

// teardownTestDB performs cleanup after tests
func teardownTestDB(t *testing.T) {
	db.CloseTestDB()
}

func TestCreateUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := authRepository.NewUserRepository(db.TestDB)

	t.Run("Valid User Creation", func(t *testing.T) {
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}

		hashedPassword := "hashed_password"

		user, err := repo.CreateUser(context.Background(), input, &hashedPassword)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
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

		if *user.HashedPassword != hashedPassword {
			t.Errorf("Expected hashed password %s, got %s", hashedPassword, *user.HashedPassword)
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
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		// Create first user
		input1 := models.CreateUserInput{
			Name:     "First User",
			Email:    "duplicate@example.com",
			Password: stringPtr("Password123!"),
		}

		hashedPassword := "hashed_password"

		_, err := repo.CreateUser(context.Background(), input1, &hashedPassword)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create second user with same email
		input2 := models.CreateUserInput{
			Name:     "Second User",
			Email:    "duplicate@example.com", // Same email
			Password: stringPtr("DifferentPass123!"),
		}

		user, err := repo.CreateUser(context.Background(), input2, &hashedPassword)

		if err == nil {
			t.Error("Expected error for duplicate email, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}
	})
}

func TestLoginUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := authRepository.NewUserRepository(db.TestDB)

	t.Run("User Not Found", func(t *testing.T) {
		loginInput := models.LoginUserInput{
			Email:    "nonexistent@example.com",
			Password: "Password123!",
		}

		user, err := repo.LoginUser(context.Background(), loginInput)

		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}

		if err != nil && err.Error() != "invalid email or password" {
			t.Errorf("Expected 'invalid email or password' error, got: %s", err.Error())
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		// Create a user first with bcrypt-hashed password
		input := models.CreateUserInput{
			Name:     "Password Test User",
			Email:    "passwordtest@example.com",
			Password: stringPtr("Password123!"),
		}

		// Generate a real bcrypt hash for testing
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		hashedPassword := string(hashedBytes)

		_, err = repo.CreateUser(context.Background(), input, &hashedPassword)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Try to login with incorrect password
		loginInput := models.LoginUserInput{
			Email:    "passwordtest@example.com",
			Password: "WrongPassword!",
		}

		user, err := repo.LoginUser(context.Background(), loginInput)

		if err == nil {
			t.Error("Expected error for wrong password, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}

		if err != nil && err.Error() != "invalid email or password" {
			t.Errorf("Expected 'invalid email or password' error, got: %s", err.Error())
		}
	})

	t.Run("Valid Login", func(t *testing.T) {
		// Create a user with bcrypt-hashed password
		input := models.CreateUserInput{
			Name:     "Valid Login User",
			Email:    "validlogin@example.com",
			Password: stringPtr("Password123!"),
		}

		// Generate a real bcrypt hash for testing
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

		// Login with correct password
		loginInput := models.LoginUserInput{
			Email:    "validlogin@example.com",
			Password: password,
		}

		user, err := repo.LoginUser(context.Background(), loginInput)

		if err != nil {
			t.Errorf("Expected successful login, got error: %v", err)
		}

		if user == nil {
			t.Error("Expected user to be returned, got nil")
		}

		if user != nil && user.Email != input.Email {
			t.Errorf("Expected email %s, got %s", input.Email, user.Email)
		}
	})
}

func TestGetUserByID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := authRepository.NewUserRepository(db.TestDB)

	// Create a user first
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

	t.Run("Existing User", func(t *testing.T) {
		user, err := repo.GetUserByID(context.Background(), createdUser.ID)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if user.ID != createdUser.ID {
			t.Errorf("Expected ID %s, got %s", createdUser.ID, user.ID)
		}

		if user.Name != input.Name {
			t.Errorf("Expected name %s, got %s", input.Name, user.Name)
		}

		if user.Email != input.Email {
			t.Errorf("Expected email %s, got %s", input.Email, user.Email)
		}
	})

	t.Run("Non-existent User", func(t *testing.T) {
		nonExistentID := uuid.New()
		user, err := repo.GetUserByID(context.Background(), nonExistentID)

		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}

		if !isPgxNoRows(err) {
			t.Errorf("Expected pgx.ErrNoRows, got: %v", err)
		}
	})
}

func TestGetUserByEmail(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := authRepository.NewUserRepository(db.TestDB)

	// Create a user first
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

	t.Run("Existing Email", func(t *testing.T) {
		user, err := repo.GetUserByEmail(context.Background(), "emailtest@example.com")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if user.Email != input.Email {
			t.Errorf("Expected email %s, got %s", input.Email, user.Email)
		}

		if user.Name != input.Name {
			t.Errorf("Expected name %s, got %s", input.Name, user.Name)
		}
	})

	t.Run("Non-existent Email", func(t *testing.T) {
		user, err := repo.GetUserByEmail(context.Background(), "nonexistent@example.com")

		if err != nil {
			t.Errorf("Expected nil error for non-existent email, got: %v", err)
		}

		if user != nil {
			t.Errorf("Expected nil user, got: %v", user)
		}
	})

	t.Run("Database Error Handling", func(t *testing.T) {
		// This would require mocking to properly test
		t.Skip("Skipping database error test - requires mocking")
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper to check if an error is pgx.ErrNoRows
func isPgxNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}

func TestGetUserWithTrips(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := authRepository.NewUserRepository(db.TestDB)
	tripRepo := tripRepository.NewTripRepository(db.TestDB)

	// Create a user first
	input := models.CreateUserInput{
		Name:     "User With Trips",
		Email:    "userwithtrips@example.com",
		Password: stringPtr("Password123!"),
	}

	hashedPassword := "hashed_password"

	user, err := repo.CreateUser(context.Background(), input, &hashedPassword)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create several trips for this user
	for i := 0; i < 5; i++ {
		tripInput := models.CreateTripInput{
			Name:        fmt.Sprintf("Trip %d", i+1),
			Description: fmt.Sprintf("Description for trip %d", i+1),
			StartDate:   time.Now().Add(time.Duration(i*24) * time.Hour),
			EndDate:     time.Now().Add(time.Duration((i+7)*24) * time.Hour),
			Destination: fmt.Sprintf("Destination %d", i+1),
		}

		_, err := tripRepo.CreateTrip(context.Background(), user.ID, tripInput)
		if err != nil {
			t.Fatalf("Failed to create test trip: %v", err)
		}
	}

	t.Run("Get User With All Trips", func(t *testing.T) {
		userWithTrips, err := repo.GetUserWithTrips(context.Background(), user.ID, 10, 0)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if userWithTrips == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if userWithTrips.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, userWithTrips.ID)
		}

		if userWithTrips.Trips == nil {
			t.Fatal("Expected trips to be populated, got nil")
		}

		if len(userWithTrips.Trips) != 5 {
			t.Errorf("Expected 5 trips, got %d", len(userWithTrips.Trips))
		}

		// Instead of checking order, just verify that all trip names exist
		tripNames := make(map[string]bool)
		for i := 1; i <= 5; i++ {
			tripNames[fmt.Sprintf("Trip %d", i)] = false
		}

		// Verify each trip belongs to the user and mark the name as found
		for _, trip := range userWithTrips.Trips {
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
	})

	t.Run("Test Pagination", func(t *testing.T) {
		// Get first page with 2 trips
		userWithTripsPage1, err := repo.GetUserWithTrips(context.Background(), user.ID, 2, 0)
		if err != nil {
			t.Fatalf("Expected no error on page 1, got: %v", err)
		}

		if len(userWithTripsPage1.Trips) != 2 {
			t.Errorf("Expected 2 trips on page 1, got %d", len(userWithTripsPage1.Trips))
		}

		// Get second page with 2 trips
		userWithTripsPage2, err := repo.GetUserWithTrips(context.Background(), user.ID, 2, 2)
		if err != nil {
			t.Fatalf("Expected no error on page 2, got: %v", err)
		}

		if len(userWithTripsPage2.Trips) != 2 {
			t.Errorf("Expected 2 trips on page 2, got %d", len(userWithTripsPage2.Trips))
		}

		// Get third page with remaining 1 trip
		userWithTripsPage3, err := repo.GetUserWithTrips(context.Background(), user.ID, 2, 4)
		if err != nil {
			t.Fatalf("Expected no error on page 3, got: %v", err)
		}

		if len(userWithTripsPage3.Trips) != 1 {
			t.Errorf("Expected 1 trip on page 3, got %d", len(userWithTripsPage3.Trips))
		}

		// Ensure all trips across pages are unique
		tripIDs := make(map[uuid.UUID]bool)

		for _, trip := range userWithTripsPage1.Trips {
			tripIDs[trip.ID] = true
		}

		for _, trip := range userWithTripsPage2.Trips {
			if tripIDs[trip.ID] {
				t.Errorf("Trip ID %s appears in multiple pages", trip.ID)
			}
			tripIDs[trip.ID] = true
		}

		for _, trip := range userWithTripsPage3.Trips {
			if tripIDs[trip.ID] {
				t.Errorf("Trip ID %s appears in multiple pages", trip.ID)
			}
			tripIDs[trip.ID] = true
		}

		if len(tripIDs) != 5 {
			t.Errorf("Expected 5 unique trip IDs across all pages, got %d", len(tripIDs))
		}
	})

	t.Run("Non-existent User", func(t *testing.T) {
		nonExistentID := uuid.New()
		userWithTrips, err := repo.GetUserWithTrips(context.Background(), nonExistentID, 10, 0)

		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if userWithTrips != nil {
			t.Errorf("Expected nil user, got: %v", userWithTrips)
		}
	})

	t.Run("User With No Trips", func(t *testing.T) {
		// Create another user with no trips
		emptyUserInput := models.CreateUserInput{
			Name:     "User With No Trips",
			Email:    "emptyuser@example.com",
			Password: stringPtr("Password123!"),
		}

		emptyUser, err := repo.CreateUser(context.Background(), emptyUserInput, &hashedPassword)
		if err != nil {
			t.Fatalf("Failed to create empty user: %v", err)
		}

		userWithTrips, err := repo.GetUserWithTrips(context.Background(), emptyUser.ID, 10, 0)
		if err != nil {
			t.Fatalf("Expected no error for user with no trips, got: %v", err)
		}

		if userWithTrips == nil {
			t.Fatal("Expected user to be returned, got nil")
		}

		if userWithTrips.Trips == nil {
			t.Fatal("Expected empty trips array, got nil")
		}

		if len(userWithTrips.Trips) != 0 {
			t.Errorf("Expected 0 trips, got %d", len(userWithTrips.Trips))
		}
	})
}
