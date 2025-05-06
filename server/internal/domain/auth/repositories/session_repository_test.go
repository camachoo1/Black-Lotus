package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/domain/auth/repositories"
	"black-lotus/internal/models"
	"black-lotus/pkg/db"
)

// Helper functions for test setup and teardown
func setupTestDB(t *testing.T) {
	if err := db.InitializeTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	if err := db.CleanTestTables(context.Background()); err != nil {
		t.Fatalf("Failed to clean test tables: %v", err)
	}
}

func teardownTestDB(t *testing.T) {
	db.CloseTestDB()
}

// Helper function to create a test user in the database
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

// Helper function to count sessions for a user
func countSessionsForUser(t *testing.T, userID uuid.UUID) int {
	var count int
	err := db.TestDB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM sessions WHERE user_id = $1
	`, userID).Scan(&count)

	if err != nil {
		t.Fatalf("Failed to count sessions: %v", err)
	}

	return count
}

func TestSessionRepository(t *testing.T) {
	t.Run("CreateSession", func(t *testing.T) {
		setupTestDB(t)
		defer teardownTestDB(t)

		repo := repositories.NewSessionRepository(db.TestDB)

		testCases := []struct {
			name            string
			setupFunc       func() uuid.UUID // returns userID
			accessDuration  time.Duration
			refreshDuration time.Duration
			expectedError   bool
			checkFunc       func(*testing.T, *models.Session, uuid.UUID, time.Duration, time.Duration)
		}{
			{
				name: "SuccessfulSessionCreation",
				setupFunc: func() uuid.UUID {
					return createTestUser(t)
				},
				accessDuration:  1 * time.Hour,
				refreshDuration: 7 * 24 * time.Hour,
				expectedError:   false,
				checkFunc: func(t *testing.T, session *models.Session, userID uuid.UUID,
					accessDuration, refreshDuration time.Duration) {
					if session.ID == uuid.Nil {
						t.Error("Expected ID to be set")
					}
					if session.UserID != userID {
						t.Errorf("Expected UserID %s, got %s", userID, session.UserID)
					}
					if session.AccessToken == "" {
						t.Error("Expected AccessToken to be set")
					}
					if session.RefreshToken == "" {
						t.Error("Expected RefreshToken to be set")
					}
					expectedAccessExpiry := time.Now().Add(accessDuration).Round(time.Second)
					if !session.AccessExpiry.Round(time.Second).Equal(expectedAccessExpiry) {
						t.Errorf("Expected AccessExpiry around %v, got %v",
							expectedAccessExpiry, session.AccessExpiry)
					}
					expectedRefreshExpiry := time.Now().Add(refreshDuration).Round(time.Second)
					if !session.RefreshExpiry.Round(time.Second).Equal(expectedRefreshExpiry) {
						t.Errorf("Expected RefreshExpiry around %v, got %v",
							expectedRefreshExpiry, session.RefreshExpiry)
					}
					if session.CreatedAt.IsZero() {
						t.Error("Expected CreatedAt to be set")
					}
				},
			},
			{
				name: "InvalidUserID",
				setupFunc: func() uuid.UUID {
					return uuid.New() // A random non-existent user ID
				},
				accessDuration:  1 * time.Hour,
				refreshDuration: 7 * 24 * time.Hour,
				expectedError:   true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				userID := tc.setupFunc()

				// Execute
				session, err := repo.CreateSession(context.Background(),
					userID, tc.accessDuration, tc.refreshDuration)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if session != nil {
						t.Errorf("Expected nil session, got: %v", session)
					}
				} else {
					if err != nil {
						t.Fatalf("Expected no error, got: %v", err)
					}
					if session == nil {
						t.Fatal("Expected session to be returned, got nil")
					}

					// Run additional checks
					if tc.checkFunc != nil {
						tc.checkFunc(t, session, userID, tc.accessDuration, tc.refreshDuration)
					}
				}
			})
		}
	})

	t.Run("TokenOperations", func(t *testing.T) {
		// Group all the token-related operations together
		tokenTests := []struct {
			name     string
			testFunc func(*testing.T)
		}{
			{
				name: "GetSessionByAccessToken",
				testFunc: func(t *testing.T) {
					setupTestDB(t)
					defer teardownTestDB(t)

					repo := repositories.NewSessionRepository(db.TestDB)
					userID := createTestUser(t)

					// Create a session first
					accessDuration := 1 * time.Hour
					refreshDuration := 7 * 24 * time.Hour
					session, err := repo.CreateSession(context.Background(),
						userID, accessDuration, refreshDuration)
					if err != nil {
						t.Fatalf("Failed to create test session: %v", err)
					}

					// Test valid token
					retrievedSession, err := repo.GetSessionByAccessToken(
						context.Background(), session.AccessToken)
					if err != nil {
						t.Errorf("Expected no error for valid token, got: %v", err)
					}
					if retrievedSession == nil {
						t.Fatal("Expected session to be returned, got nil")
					}
					if retrievedSession.ID != session.ID {
						t.Errorf("Expected ID %s, got %s", session.ID, retrievedSession.ID)
					}
					if retrievedSession.UserID != userID {
						t.Errorf("Expected UserID %s, got %s", userID, retrievedSession.UserID)
					}

					// Test invalid token
					invalidToken := "invalid_token"
					retrievedSession, err = repo.GetSessionByAccessToken(
						context.Background(), invalidToken)
					if err == nil {
						t.Error("Expected error for invalid token, got nil")
					}
					if retrievedSession != nil {
						t.Errorf("Expected nil session, got: %v", retrievedSession)
					}
				},
			},
			{
				name: "GetSessionByRefreshToken",
				testFunc: func(t *testing.T) {
					setupTestDB(t)
					defer teardownTestDB(t)

					repo := repositories.NewSessionRepository(db.TestDB)
					userID := createTestUser(t)

					// Create a session first
					accessDuration := 1 * time.Hour
					refreshDuration := 7 * 24 * time.Hour
					session, err := repo.CreateSession(context.Background(),
						userID, accessDuration, refreshDuration)
					if err != nil {
						t.Fatalf("Failed to create test session: %v", err)
					}

					// Test valid token
					retrievedSession, err := repo.GetSessionByRefreshToken(
						context.Background(), session.RefreshToken)
					if err != nil {
						t.Errorf("Expected no error for valid token, got: %v", err)
					}
					if retrievedSession == nil {
						t.Fatal("Expected session to be returned, got nil")
					}
					if retrievedSession.ID != session.ID {
						t.Errorf("Expected ID %s, got %s", session.ID, retrievedSession.ID)
					}
					if retrievedSession.UserID != userID {
						t.Errorf("Expected UserID %s, got %s", userID, retrievedSession.UserID)
					}

					// Test invalid token
					invalidToken := "invalid_token"
					retrievedSession, err = repo.GetSessionByRefreshToken(
						context.Background(), invalidToken)
					if err == nil {
						t.Error("Expected error for invalid token, got nil")
					}
					if retrievedSession != nil {
						t.Errorf("Expected nil session, got: %v", retrievedSession)
					}
				},
			},
			{
				name: "RefreshAccessToken",
				testFunc: func(t *testing.T) {
					setupTestDB(t)
					defer teardownTestDB(t)

					repo := repositories.NewSessionRepository(db.TestDB)
					userID := createTestUser(t)

					// Create a session first
					accessDuration := 1 * time.Hour
					refreshDuration := 7 * 24 * time.Hour
					session, err := repo.CreateSession(context.Background(),
						userID, accessDuration, refreshDuration)
					if err != nil {
						t.Fatalf("Failed to create test session: %v", err)
					}

					// Store original values for comparison
					originalAccessToken := session.AccessToken
					originalAccessExpiry := session.AccessExpiry
					originalRefreshExpiry := session.RefreshExpiry

					// Test refreshing the access token
					updatedSession, err := repo.RefreshAccessToken(context.Background(), session.ID)
					if err != nil {
						t.Errorf("Expected no error for valid session ID, got: %v", err)
					}
					if updatedSession == nil {
						t.Fatal("Expected session to be returned, got nil")
					}
					if updatedSession.ID != session.ID {
						t.Errorf("Expected ID %s, got %s", session.ID, updatedSession.ID)
					}
					if updatedSession.AccessToken == originalAccessToken {
						t.Error("Expected new access token to be different from original")
					}
					if updatedSession.AccessToken == "" {
						t.Error("Expected new access token to be set")
					}
					if !updatedSession.AccessExpiry.After(originalAccessExpiry) {
						t.Error("Expected new access expiry to be after original expiry")
					}
					if !updatedSession.RefreshExpiry.Equal(originalRefreshExpiry) {
						t.Error("Expected refresh expiry to remain unchanged")
					}

					// Test invalid session ID
					invalidID := uuid.New()
					updatedSession, err = repo.RefreshAccessToken(context.Background(), invalidID)
					if err == nil {
						t.Error("Expected error for invalid session ID, got nil")
					}
					if updatedSession != nil {
						t.Errorf("Expected nil session, got: %v", updatedSession)
					}
				},
			},
		}

		for _, tt := range tokenTests {
			t.Run(tt.name, func(t *testing.T) {
				tt.testFunc(t)
			})
		}
	})

	t.Run("DeleteOperations", func(t *testing.T) {
		// Group all delete operations together
		deleteTests := []struct {
			name     string
			testFunc func(*testing.T)
		}{
			{
				name: "DeleteSessionByAccessToken",
				testFunc: func(t *testing.T) {
					setupTestDB(t)
					defer teardownTestDB(t)

					repo := repositories.NewSessionRepository(db.TestDB)
					userID := createTestUser(t)

					// Create a session first
					accessDuration := 1 * time.Hour
					refreshDuration := 7 * 24 * time.Hour
					session, err := repo.CreateSession(context.Background(),
						userID, accessDuration, refreshDuration)
					if err != nil {
						t.Fatalf("Failed to create test session: %v", err)
					}

					// Delete the session by access token
					err = repo.DeleteSessionByAccessToken(context.Background(), session.AccessToken)
					if err != nil {
						t.Errorf("Expected no error when deleting, got: %v", err)
					}

					// Try to retrieve the deleted session
					retrievedSession, err := repo.GetSessionByAccessToken(
						context.Background(), session.AccessToken)
					if err == nil {
						t.Error("Expected error after deletion, got nil")
					}
					if retrievedSession != nil {
						t.Errorf("Expected nil session after deletion, got: %v", retrievedSession)
					}

					// Test with invalid token
					invalidToken := "invalid_token"
					err = repo.DeleteSessionByAccessToken(context.Background(), invalidToken)
					if err != nil {
						t.Errorf("Expected no error for non-existent token, got: %v", err)
					}
				},
			},
			{
				name: "DeleteSessionByRefreshToken",
				testFunc: func(t *testing.T) {
					setupTestDB(t)
					defer teardownTestDB(t)

					repo := repositories.NewSessionRepository(db.TestDB)
					userID := createTestUser(t)

					// Create a session first
					accessDuration := 1 * time.Hour
					refreshDuration := 7 * 24 * time.Hour
					session, err := repo.CreateSession(context.Background(),
						userID, accessDuration, refreshDuration)
					if err != nil {
						t.Fatalf("Failed to create test session: %v", err)
					}

					// Delete the session by refresh token
					err = repo.DeleteSessionByRefreshToken(context.Background(), session.RefreshToken)
					if err != nil {
						t.Errorf("Expected no error when deleting, got: %v", err)
					}

					// Try to retrieve the deleted session
					retrievedSession, err := repo.GetSessionByRefreshToken(
						context.Background(), session.RefreshToken)
					if err == nil {
						t.Error("Expected error after deletion, got nil")
					}
					if retrievedSession != nil {
						t.Errorf("Expected nil session after deletion, got: %v", retrievedSession)
					}

					// Test with invalid token
					invalidToken := "invalid_token"
					err = repo.DeleteSessionByRefreshToken(context.Background(), invalidToken)
					if err != nil {
						t.Errorf("Expected no error for non-existent token, got: %v", err)
					}
				},
			},
			{
				name: "DeleteUserSessions",
				testFunc: func(t *testing.T) {
					setupTestDB(t)
					defer teardownTestDB(t)

					repo := repositories.NewSessionRepository(db.TestDB)
					userID := createTestUser(t)

					// Create multiple sessions for the user
					accessDuration := 1 * time.Hour
					refreshDuration := 7 * 24 * time.Hour
					for i := 0; i < 3; i++ {
						_, err := repo.CreateSession(context.Background(),
							userID, accessDuration, refreshDuration)
						if err != nil {
							t.Fatalf("Failed to create test session %d: %v", i, err)
						}
					}

					// Verify sessions were created
					count := countSessionsForUser(t, userID)
					if count != 3 {
						t.Errorf("Expected 3 sessions before deletion, found %d", count)
					}

					// Delete all sessions for the user
					err := repo.DeleteUserSessions(context.Background(), userID)
					if err != nil {
						t.Errorf("Expected no error when deleting, got: %v", err)
					}

					// Verify all sessions were deleted
					count = countSessionsForUser(t, userID)
					if count != 0 {
						t.Errorf("Expected 0 sessions after deletion, found %d", count)
					}

					// Test with user that has no sessions
					nonExistentUserID := uuid.New()
					err = repo.DeleteUserSessions(context.Background(), nonExistentUserID)
					if err != nil {
						t.Errorf("Expected no error for user with no sessions, got: %v", err)
					}
				},
			},
		}

		for _, tt := range deleteTests {
			t.Run(tt.name, func(t *testing.T) {
				tt.testFunc(t)
			})
		}
	})
}
