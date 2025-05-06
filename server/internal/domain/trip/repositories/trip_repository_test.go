package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/domain/trip/repositories"
	"black-lotus/internal/models"
	"black-lotus/pkg/db"
)

// setupTestDB ensures the test database is initialized
func setupTestDB(t *testing.T) {
	if err := db.InitializeTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Clean test tables to ensure a fresh state
	if err := db.CleanTestTables(context.Background()); err != nil {
		t.Fatalf("Failed to clean test tables: %v", err)
	}
}

// createTestUser creates a user in the test database for testing
func createTestUser(t *testing.T) uuid.UUID {
	userID := uuid.New()

	_, err := db.TestDB.Exec(context.Background(), `
        INSERT INTO users (id, name, email, hashed_password, email_verified)
        VALUES ($1, 'Test User', 'test@example.com', 'password', true)
    `, userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// teardownTestDB performs cleanup after tests
func teardownTestDB(t *testing.T) {
	db.CloseTestDB()
}

func TestCreateTrip(t *testing.T) {
	setupTestDB(t)

	defer teardownTestDB(t)

	userID := createTestUser(t)
	repo := repositories.NewTripRepository(db.TestDB)

	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(7 * 24 * time.Hour)

	// Test case: Valid trip creation
	t.Run("Valid Trip", func(t *testing.T) {
		input := models.CreateTripInput{
			Name:        "Test Trip",
			Description: "A test trip",
			StartDate:   startDate,
			EndDate:     endDate,
			Destination: "Test City",
		}

		trip, err := repo.CreateTrip(context.Background(), userID, input)
		if err != nil {
			t.Fatalf("Should create trip without error: %v", err)
		}

		if trip == nil {
			t.Fatal("Trip should not be nil")
		}

		if trip.ID == uuid.Nil {
			t.Error("Trip ID should be set")
		}

		if trip.UserID != userID {
			t.Errorf("Trip should be associated with user ID %s, got %s", userID, trip.UserID)
		}

		if trip.Name != input.Name {
			t.Errorf("Trip name should be %s, got %s", input.Name, trip.Name)
		}

		if trip.Description != input.Description {
			t.Errorf("Trip description should be %s, got %s", input.Description, trip.Description)
		}

		if !trip.StartDate.Equal(input.StartDate) {
			t.Errorf("Start date should be %v, got %v", input.StartDate, trip.StartDate)
		}

		if !trip.EndDate.Equal(input.EndDate) {
			t.Errorf("End date should be %v, got %v", input.EndDate, trip.EndDate)
		}

		if trip.Destination != input.Destination {
			t.Errorf("Destination should be %s, got %s", input.Destination, trip.Destination)
		}

		if trip.CreatedAt.IsZero() {
			t.Error("Created time should be set")
		}

		if trip.UpdatedAt.IsZero() {
			t.Error("Updated time should be set")
		}
	})
}

func TestGetTripByID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t)
	repo := repositories.NewTripRepository(db.TestDB)

	// First create a trip to test with
	input := models.CreateTripInput{
		Name:        "Test Trip",
		Description: "A test trip",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(7 * 24 * time.Hour),
		Destination: "Test City",
	}

	createdTrip, err := repo.CreateTrip(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("Should create trip without error: %v", err)
	}

	// Now test fetching the trip we just created
	t.Run("Existing Trip", func(t *testing.T) {
		trip, err := repo.GetTripByID(context.Background(), createdTrip.ID)
		if err != nil {
			t.Fatalf("Should get trip without error: %v", err)
		}

		if trip.ID != createdTrip.ID {
			t.Errorf("Trip ID should be %s, got %s", createdTrip.ID, trip.ID)
		}

		if trip.UserID != userID {
			t.Errorf("User ID should be %s, got %s", userID, trip.UserID)
		}

		if trip.Name != input.Name {
			t.Errorf("Trip name should be %s, got %s", input.Name, trip.Name)
		}
	})

	// Test case: Non-existent trip
	t.Run("Non-existent Trip", func(t *testing.T) {
		nonExistentID := uuid.New()
		trip, err := repo.GetTripByID(context.Background(), nonExistentID)

		if err == nil {
			t.Error("Should return error for non-existent trip")
		}

		if trip != nil {
			t.Error("Trip should be nil")
		}
	})
}

func TestUpdateTrip(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t)
	repo := repositories.NewTripRepository(db.TestDB)

	// First create a trip to test with
	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(7 * 24 * time.Hour)

	input := models.CreateTripInput{
		Name:        "Test Trip",
		Description: "A test trip",
		StartDate:   startDate,
		EndDate:     endDate,
		Destination: "Test City",
	}

	createdTrip, err := repo.CreateTrip(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("Should create trip without error: %v", err)
	}

	// Helper functions for creating pointers
	strPtr := func(s string) *string { return &s }

	// Test case: Update trip name
	t.Run("Update Name", func(t *testing.T) {
		newName := "Updated Trip Name"
		updateInput := models.UpdateTripInput{
			Name: strPtr(newName),
		}

		updatedTrip, err := repo.UpdateTrip(context.Background(), createdTrip.ID, updateInput)
		if err != nil {
			t.Fatalf("Should update trip without error: %v", err)
		}

		if updatedTrip.Name != newName {
			t.Errorf("Trip name should be updated to %s, got %s", newName, updatedTrip.Name)
		}

		if updatedTrip.Description != input.Description {
			t.Errorf("Description should remain unchanged as %s, got %s", input.Description, updatedTrip.Description)
		}
	})

	// Test case: Update non-existent trip
	t.Run("Update Non-existent Trip", func(t *testing.T) {
		nonExistentID := uuid.New()
		updateInput := models.UpdateTripInput{
			Name: strPtr("This won't work"),
		}

		updatedTrip, err := repo.UpdateTrip(context.Background(), nonExistentID, updateInput)
		if err == nil {
			t.Error("Should return error for non-existent trip")
		}

		if updatedTrip != nil {
			t.Error("Updated trip should be nil")
		}
	})
}

func TestDeleteTrip(t *testing.T) {
	setupTestDB(t)

	defer teardownTestDB(t)

	userID := createTestUser(t)
	repo := repositories.NewTripRepository(db.TestDB)

	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(7 * 24 * time.Hour)

	// Create a trip for testing
	input := models.CreateTripInput{
		Name:        "Test Trip",
		Description: "A test trip",
		StartDate:   startDate,
		EndDate:     endDate,
		Destination: "Test City",
	}

	createdTrip, err := repo.CreateTrip(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("Should create trip without error: %v", err)
	}

	// Test case: Delete existing trip
	t.Run("Delete Existing Trip", func(t *testing.T) {
		err := repo.DeleteTrip(context.Background(), createdTrip.ID)
		if err != nil {
			t.Fatalf("Should delete trip without error: %v", err)
		}

		// Verify trip is deleted
		trip, err := repo.GetTripByID(context.Background(), createdTrip.ID)
		if err == nil {
			t.Error("Should return error when getting deleted trip")
		}

		if trip != nil {
			t.Error("Trip should be nil after deletion")
		}
	})

	// Test case: Delete non-existent trip
	t.Run("Delete Non-existent Trip", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := repo.DeleteTrip(context.Background(), nonExistentID)

		if err == nil {
			t.Error("Should return error for non-existent trip")
		}

		if err != nil && err.Error() != "trip not found" {
			t.Errorf("Error should indicate trip not found, got: %v", err)
		}
	})
}

func TestGetTripsByUserID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t)
	anotherUserID := uuid.New()
	repo := repositories.NewTripRepository(db.TestDB)

	// Create multiple trips for testing
	for i := 0; i < 5; i++ {
		startDate := time.Now().Add(time.Duration(i*24) * time.Hour)
		endDate := startDate.Add(7 * 24 * time.Hour)

		input := models.CreateTripInput{
			Name:        "Test Trip",
			Description: "Test Description",
			StartDate:   startDate,
			EndDate:     endDate,
			Destination: "Test City",
		}

		_, err := repo.CreateTrip(context.Background(), userID, input)
		if err != nil {
			t.Fatalf("Should create trip without error: %v", err)
		}
	}

	// Test case: Get all trips for user
	t.Run("Get All User Trips", func(t *testing.T) {
		trips, err := repo.GetTripsByUserID(context.Background(), userID, 10, 0)
		if err != nil {
			t.Fatalf("Should get trips without error: %v", err)
		}

		if len(trips) != 5 {
			t.Errorf("Should return 5 trips, got %d", len(trips))
		}

		// Check that all trips belong to the specified user
		for _, trip := range trips {
			if trip.UserID != userID {
				t.Errorf("Trip should belong to user %s, got %s", userID, trip.UserID)
			}
		}
	})

	// Test case: Test pagination
	t.Run("Test Pagination", func(t *testing.T) {
		// Get first page (2 trips)
		page1, err := repo.GetTripsByUserID(context.Background(), userID, 2, 0)
		if err != nil {
			t.Fatalf("Should get first page without error: %v", err)
		}

		if len(page1) != 2 {
			t.Errorf("Should return 2 trips, got %d", len(page1))
		}

		// Get second page (2 trips)
		page2, err := repo.GetTripsByUserID(context.Background(), userID, 2, 2)
		if err != nil {
			t.Fatalf("Should get second page without error: %v", err)
		}

		if len(page2) != 2 {
			t.Errorf("Should return 2 trips, got %d", len(page2))
		}

		// Get third page (1 trip)
		page3, err := repo.GetTripsByUserID(context.Background(), userID, 2, 4)
		if err != nil {
			t.Fatalf("Should get third page without error: %v", err)
		}

		if len(page3) != 1 {
			t.Errorf("Should return 1 trip, got %d", len(page3))
		}

		// Ensure all IDs are different (no duplicates across pages)
		allIDs := make(map[uuid.UUID]bool)
		for _, trip := range append(append(page1, page2...), page3...) {
			if allIDs[trip.ID] {
				t.Errorf("Trip ID %s should not appear more than once", trip.ID)
			}
			allIDs[trip.ID] = true
		}

		if len(allIDs) != 5 {
			t.Errorf("Should have 5 unique trips across all pages, got %d", len(allIDs))
		}
	})

	// Test case: No trips for a user
	t.Run("No Trips", func(t *testing.T) {
		trips, err := repo.GetTripsByUserID(context.Background(), anotherUserID, 10, 0)
		if err != nil {
			t.Fatalf("Should not error when no trips exist: %v", err)
		}

		if len(trips) != 0 {
			t.Errorf("Should return empty slice, got %d trips", len(trips))
		}
	})
}

func TestGetTripWithUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t)
	repo := repositories.NewTripRepository(db.TestDB)

	// First create a trip to test with
	input := models.CreateTripInput{
		Name:        "Test Trip",
		Description: "A test trip",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(7 * 24 * time.Hour),
		Destination: "Test City",
	}

	createdTrip, err := repo.CreateTrip(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("Should create trip without error: %v", err)
	}

	// Test case: Get trip with user
	t.Run("Get Trip With User", func(t *testing.T) {
		trip, err := repo.GetTripWithUser(context.Background(), createdTrip.ID)
		if err != nil {
			t.Fatalf("Should get trip with user without error: %v", err)
		}

		if trip == nil {
			t.Fatal("Trip should not be nil")
		}

		if trip.ID != createdTrip.ID {
			t.Errorf("Trip ID should be %s, got %s", createdTrip.ID, trip.ID)
		}

		if trip.UserID != userID {
			t.Errorf("User ID should be %s, got %s", userID, trip.UserID)
		}

		// Verify the User field is populated
		if trip.User == nil {
			t.Error("User field should be populated")
		}

		if trip.User != nil && trip.User.ID != userID {
			t.Errorf("User ID in User field should be %s, got %s", userID, trip.User.ID)
		}

		if trip.User != nil && trip.User.Name != "Test User" {
			t.Errorf("User name should be 'Test User', got %s", trip.User.Name)
		}
	})

	// Test case: Non-existent trip
	t.Run("Non-existent Trip With User", func(t *testing.T) {
		nonExistentID := uuid.New()
		trip, err := repo.GetTripWithUser(context.Background(), nonExistentID)

		if err == nil {
			t.Error("Should return error for non-existent trip")
		}

		if trip != nil {
			t.Error("Trip should be nil")
		}
	})
}
