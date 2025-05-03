package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/db"
	"black-lotus/internal/auth/repositories"
)

// createTestUser creates a user for testing
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

func TestCreateSession(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	t.Run("Successful Session Creation", func(t *testing.T) {
		accessDuration := 1 * time.Hour
		refreshDuration := 7 * 24 * time.Hour

		session, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

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
			t.Errorf("Expected AccessExpiry around %v, got %v", expectedAccessExpiry, session.AccessExpiry)
		}

		expectedRefreshExpiry := time.Now().Add(refreshDuration).Round(time.Second)
		if !session.RefreshExpiry.Round(time.Second).Equal(expectedRefreshExpiry) {
			t.Errorf("Expected RefreshExpiry around %v, got %v", expectedRefreshExpiry, session.RefreshExpiry)
		}

		if session.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("Invalid User ID", func(t *testing.T) {
		nonExistentUserID := uuid.New()
		accessDuration := 1 * time.Hour
		refreshDuration := 7 * 24 * time.Hour

		session, err := repo.CreateSession(context.Background(), nonExistentUserID, accessDuration, refreshDuration)

		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}

		if session != nil {
			t.Errorf("Expected nil session, got: %v", session)
		}
	})
}

func TestGetSessionByAccessToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	// Create a session first
	accessDuration := 1 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	session, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Valid Token", func(t *testing.T) {
		retrievedSession, err := repo.GetSessionByAccessToken(context.Background(), session.AccessToken)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
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
	})

	t.Run("Invalid Token", func(t *testing.T) {
		invalidToken := "invalid_token"
		retrievedSession, err := repo.GetSessionByAccessToken(context.Background(), invalidToken)

		if err == nil {
			t.Error("Expected error for invalid token, got nil")
		}

		if retrievedSession != nil {
			t.Errorf("Expected nil session, got: %v", retrievedSession)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		// This would need to manipulate the database directly to set an expired token
		// For now, we'll skip this test as it requires more complex setup
		t.Skip("Skipping expired token test - requires database manipulation")
	})
}

func TestGetSessionByRefreshToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	// Create a session first
	accessDuration := 1 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	session, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Valid Token", func(t *testing.T) {
		retrievedSession, err := repo.GetSessionByRefreshToken(context.Background(), session.RefreshToken)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
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
	})

	t.Run("Invalid Token", func(t *testing.T) {
		invalidToken := "invalid_token"
		retrievedSession, err := repo.GetSessionByRefreshToken(context.Background(), invalidToken)

		if err == nil {
			t.Error("Expected error for invalid token, got nil")
		}

		if retrievedSession != nil {
			t.Errorf("Expected nil session, got: %v", retrievedSession)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		// This would need to manipulate the database directly to set an expired token
		// For now, we'll skip this test as it requires more complex setup
		t.Skip("Skipping expired token test - requires database manipulation")
	})
}

func TestRefreshAccessToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	// Create a session first
	accessDuration := 1 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	session, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Valid Session ID", func(t *testing.T) {
		updatedSession, err := repo.RefreshAccessToken(context.Background(), session.ID)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if updatedSession == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if updatedSession.ID != session.ID {
			t.Errorf("Expected ID %s, got %s", session.ID, updatedSession.ID)
		}

		if updatedSession.AccessToken == session.AccessToken {
			t.Error("Expected new access token to be different from original")
		}

		if updatedSession.AccessToken == "" {
			t.Error("Expected new access token to be set")
		}

		// Access expiry should be updated
		if !updatedSession.AccessExpiry.After(session.AccessExpiry) {
			t.Error("Expected new access expiry to be after original expiry")
		}

		// Refresh expiry should remain the same
		if !updatedSession.RefreshExpiry.Equal(session.RefreshExpiry) {
			t.Error("Expected refresh expiry to remain unchanged")
		}
	})

	t.Run("Invalid Session ID", func(t *testing.T) {
		invalidID := uuid.New()
		updatedSession, err := repo.RefreshAccessToken(context.Background(), invalidID)

		if err == nil {
			t.Error("Expected error for invalid session ID, got nil")
		}

		if updatedSession != nil {
			t.Errorf("Expected nil session, got: %v", updatedSession)
		}
	})
}

func TestDeleteSessionByAccessToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	// Create a session first
	accessDuration := 1 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	session, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Valid Token", func(t *testing.T) {
		// Delete the session
		err := repo.DeleteSessionByAccessToken(context.Background(), session.AccessToken)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Try to retrieve the deleted session
		retrievedSession, err := repo.GetSessionByAccessToken(context.Background(), session.AccessToken)

		if err == nil {
			t.Error("Expected error after deletion, got nil")
		}

		if retrievedSession != nil {
			t.Errorf("Expected nil session after deletion, got: %v", retrievedSession)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		invalidToken := "invalid_token"
		err := repo.DeleteSessionByAccessToken(context.Background(), invalidToken)

		// The function doesn't return an error if the token doesn't exist
		if err != nil {
			t.Errorf("Expected no error for non-existent token, got: %v", err)
		}
	})
}

func TestDeleteSessionByRefreshToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	// Create a session first
	accessDuration := 1 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	session, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Valid Token", func(t *testing.T) {
		// Delete the session
		err := repo.DeleteSessionByRefreshToken(context.Background(), session.RefreshToken)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Try to retrieve the deleted session
		retrievedSession, err := repo.GetSessionByRefreshToken(context.Background(), session.RefreshToken)

		if err == nil {
			t.Error("Expected error after deletion, got nil")
		}

		if retrievedSession != nil {
			t.Errorf("Expected nil session after deletion, got: %v", retrievedSession)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		invalidToken := "invalid_token"
		err := repo.DeleteSessionByRefreshToken(context.Background(), invalidToken)

		// The function doesn't return an error if the token doesn't exist
		if err != nil {
			t.Errorf("Expected no error for non-existent token, got: %v", err)
		}
	})
}

func TestDeleteUserSessions(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	repo := repositories.NewSessionRepository(db.TestDB)
	userID := createTestUser(t)

	// Create multiple sessions for the user
	accessDuration := 1 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	for i := 0; i < 3; i++ {
		_, err := repo.CreateSession(context.Background(), userID, accessDuration, refreshDuration)
		if err != nil {
			t.Fatalf("Failed to create test session %d: %v", i, err)
		}
	}

	t.Run("Delete All User Sessions", func(t *testing.T) {
		// Delete all sessions for the user
		err := repo.DeleteUserSessions(context.Background(), userID)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify that all sessions are deleted
		// We'd need to add a method to count sessions for a user, but for now we'll just check
		// that we can't find any of the sessions we created
		count := countSessionsForUser(t, userID)
		if count != 0 {
			t.Errorf("Expected 0 sessions after deletion, found %d", count)
		}
	})

	t.Run("User Without Sessions", func(t *testing.T) {
		nonExistentUserID := uuid.New()
		err := repo.DeleteUserSessions(context.Background(), nonExistentUserID)

		// The function doesn't return an error if there are no sessions for the user
		if err != nil {
			t.Errorf("Expected no error for user with no sessions, got: %v", err)
		}
	})
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
