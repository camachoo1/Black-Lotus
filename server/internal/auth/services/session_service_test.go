package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/auth/models"
	"black-lotus/internal/auth/services"
)

// MockSessionRepository is a mock implementation of SessionRepository for testing
type MockSessionRepository struct {
	createSessionFunc               func(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error)
	getSessionByAccessTokenFunc     func(ctx context.Context, token string) (*models.Session, error)
	getSessionByRefreshTokenFunc    func(ctx context.Context, token string) (*models.Session, error)
	refreshAccessTokenFunc          func(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)
	deleteSessionByAccessTokenFunc  func(ctx context.Context, token string) error
	deleteSessionByRefreshTokenFunc func(ctx context.Context, token string) error
	deleteUserSessionsFunc          func(ctx context.Context, userID uuid.UUID) error
}

// CreateSession mocks the repository's CreateSession method
func (m *MockSessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, userID, accessDuration, refreshDuration)
	}
	return &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   "mock_access_token",
		RefreshToken:  "mock_refresh_token",
		AccessExpiry:  time.Now().Add(accessDuration),
		RefreshExpiry: time.Now().Add(refreshDuration),
		CreatedAt:     time.Now(),
	}, nil
}

// GetSessionByAccessToken mocks the repository's GetSessionByAccessToken method
func (m *MockSessionRepository) GetSessionByAccessToken(ctx context.Context, token string) (*models.Session, error) {
	if m.getSessionByAccessTokenFunc != nil {
		return m.getSessionByAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("token not found")
}

// GetSessionByRefreshToken mocks the repository's GetSessionByRefreshToken method
func (m *MockSessionRepository) GetSessionByRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	if m.getSessionByRefreshTokenFunc != nil {
		return m.getSessionByRefreshTokenFunc(ctx, token)
	}
	return nil, errors.New("token not found")
}

// RefreshAccessToken mocks the repository's RefreshAccessToken method
func (m *MockSessionRepository) RefreshAccessToken(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	if m.refreshAccessTokenFunc != nil {
		return m.refreshAccessTokenFunc(ctx, sessionID)
	}
	return &models.Session{
		ID:            sessionID,
		UserID:        uuid.New(),
		AccessToken:   "new_mock_access_token",
		RefreshToken:  "mock_refresh_token",
		AccessExpiry:  time.Now().Add(time.Hour),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:     time.Now(),
	}, nil
}

// DeleteSessionByAccessToken mocks the repository's DeleteSessionByAccessToken method
func (m *MockSessionRepository) DeleteSessionByAccessToken(ctx context.Context, token string) error {
	if m.deleteSessionByAccessTokenFunc != nil {
		return m.deleteSessionByAccessTokenFunc(ctx, token)
	}
	return nil
}

// DeleteSessionByRefreshToken mocks the repository's DeleteSessionByRefreshToken method
func (m *MockSessionRepository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	if m.deleteSessionByRefreshTokenFunc != nil {
		return m.deleteSessionByRefreshTokenFunc(ctx, token)
	}
	return nil
}

// DeleteUserSessions mocks the repository's DeleteUserSessions method
func (m *MockSessionRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	if m.deleteUserSessionsFunc != nil {
		return m.deleteUserSessionsFunc(ctx, userID)
	}
	return nil
}

func TestCreateSession(t *testing.T) {
	t.Run("Create New Session", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			createSessionFunc: func(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
				if accessDuration != time.Hour {
					t.Errorf("Expected access duration of 1 hour, got %v", accessDuration)
				}
				if refreshDuration != 7*24*time.Hour {
					t.Errorf("Expected refresh duration of 7 days, got %v", refreshDuration)
				}
				return &models.Session{
					ID:            uuid.New(),
					UserID:        userID,
					AccessToken:   "test_access_token",
					RefreshToken:  "test_refresh_token",
					AccessExpiry:  time.Now().Add(accessDuration),
					RefreshExpiry: time.Now().Add(refreshDuration),
					CreatedAt:     time.Now(),
				}, nil
			},
		}
		service := services.NewSessionService(mockRepo)

		// Input
		userID := uuid.New()

		// Execute
		session, err := service.CreateSession(context.Background(), userID)

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
		}

		if session.AccessToken != "test_access_token" {
			t.Errorf("Expected access token 'test_access_token', got '%s'", session.AccessToken)
		}

		if session.RefreshToken != "test_refresh_token" {
			t.Errorf("Expected refresh token 'test_refresh_token', got '%s'", session.RefreshToken)
		}
	})

	t.Run("Repository Error", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			createSessionFunc: func(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
				return nil, errors.New("database error")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Input
		userID := uuid.New()

		// Execute
		session, err := service.CreateSession(context.Background(), userID)

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if session != nil {
			t.Errorf("Expected nil session, got: %v", session)
		}

		if err.Error() != "database error" {
			t.Errorf("Expected error message 'database error', got '%s'", err.Error())
		}
	})
}

func TestValidateAccessToken(t *testing.T) {
	t.Run("Valid Access Token", func(t *testing.T) {
		// Setup
		userID := uuid.New()
		validSession := &models.Session{
			ID:            uuid.New(),
			UserID:        userID,
			AccessToken:   "valid_access_token",
			RefreshToken:  "valid_refresh_token",
			AccessExpiry:  time.Now().Add(time.Hour),
			RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo := &MockSessionRepository{
			getSessionByAccessTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				if token == "valid_access_token" {
					return validSession, nil
				}
				return nil, errors.New("invalid token")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.ValidateAccessToken(context.Background(), "valid_access_token")

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
		}
	})

	t.Run("Invalid Access Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			getSessionByAccessTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				return nil, errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.ValidateAccessToken(context.Background(), "invalid_token")

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if session != nil {
			t.Errorf("Expected nil session, got: %v", session)
		}

		if err.Error() != "token not found" {
			t.Errorf("Expected error message 'token not found', got '%s'", err.Error())
		}
	})
}

func TestValidateRefreshToken(t *testing.T) {
	t.Run("Valid Refresh Token", func(t *testing.T) {
		// Setup
		userID := uuid.New()
		validSession := &models.Session{
			ID:            uuid.New(),
			UserID:        userID,
			AccessToken:   "valid_access_token",
			RefreshToken:  "valid_refresh_token",
			AccessExpiry:  time.Now().Add(time.Hour),
			RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo := &MockSessionRepository{
			getSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				if token == "valid_refresh_token" {
					return validSession, nil
				}
				return nil, errors.New("invalid token")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.ValidateRefreshToken(context.Background(), "valid_refresh_token")

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
		}
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			getSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				return nil, errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.ValidateRefreshToken(context.Background(), "invalid_token")

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if session != nil {
			t.Errorf("Expected nil session, got: %v", session)
		}

		if err.Error() != "token not found" {
			t.Errorf("Expected error message 'token not found', got '%s'", err.Error())
		}
	})
}

func TestRefreshAccessToken(t *testing.T) {
	t.Run("Valid Refresh Token", func(t *testing.T) {
		// Setup
		sessionID := uuid.New()
		userID := uuid.New()
		validSession := &models.Session{
			ID:            sessionID,
			UserID:        userID,
			AccessToken:   "old_access_token",
			RefreshToken:  "valid_refresh_token",
			AccessExpiry:  time.Now().Add(time.Hour),
			RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		}

		newSession := &models.Session{
			ID:            sessionID,
			UserID:        userID,
			AccessToken:   "new_access_token",
			RefreshToken:  "valid_refresh_token",
			AccessExpiry:  time.Now().Add(time.Hour),
			RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo := &MockSessionRepository{
			getSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				if token == "valid_refresh_token" {
					return validSession, nil
				}
				return nil, errors.New("invalid token")
			},
			refreshAccessTokenFunc: func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
				if id == sessionID {
					return newSession, nil
				}
				return nil, errors.New("session not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.RefreshAccessToken(context.Background(), "valid_refresh_token")

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.AccessToken != "new_access_token" {
			t.Errorf("Expected new access token 'new_access_token', got '%s'", session.AccessToken)
		}
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			getSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				return nil, errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.RefreshAccessToken(context.Background(), "invalid_token")

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if session != nil {
			t.Errorf("Expected nil session, got: %v", session)
		}

		if err.Error() != "token not found" {
			t.Errorf("Expected error message 'token not found', got '%s'", err.Error())
		}
	})

	t.Run("Refresh Error", func(t *testing.T) {
		// Setup
		sessionID := uuid.New()
		userID := uuid.New()
		validSession := &models.Session{
			ID:            sessionID,
			UserID:        userID,
			AccessToken:   "old_access_token",
			RefreshToken:  "valid_refresh_token",
			AccessExpiry:  time.Now().Add(time.Hour),
			RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo := &MockSessionRepository{
			getSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.Session, error) {
				return validSession, nil
			},
			refreshAccessTokenFunc: func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
				return nil, errors.New("database error")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		session, err := service.RefreshAccessToken(context.Background(), "valid_refresh_token")

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if session != nil {
			t.Errorf("Expected nil session, got: %v", session)
		}

		if err.Error() != "database error" {
			t.Errorf("Expected error message 'database error', got '%s'", err.Error())
		}
	})
}

func TestEndSessionByAccessToken(t *testing.T) {
	t.Run("Valid Access Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			deleteSessionByAccessTokenFunc: func(ctx context.Context, token string) error {
				if token == "valid_access_token" {
					return nil
				}
				return errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		err := service.EndSessionByAccessToken(context.Background(), "valid_access_token")

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid Access Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			deleteSessionByAccessTokenFunc: func(ctx context.Context, token string) error {
				return errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		err := service.EndSessionByAccessToken(context.Background(), "invalid_token")

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "token not found" {
			t.Errorf("Expected error message 'token not found', got '%s'", err.Error())
		}
	})
}

func TestEndSessionByRefreshToken(t *testing.T) {
	t.Run("Valid Refresh Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			deleteSessionByRefreshTokenFunc: func(ctx context.Context, token string) error {
				if token == "valid_refresh_token" {
					return nil
				}
				return errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		err := service.EndSessionByRefreshToken(context.Background(), "valid_refresh_token")

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		// Setup
		mockRepo := &MockSessionRepository{
			deleteSessionByRefreshTokenFunc: func(ctx context.Context, token string) error {
				return errors.New("token not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		err := service.EndSessionByRefreshToken(context.Background(), "invalid_token")

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "token not found" {
			t.Errorf("Expected error message 'token not found', got '%s'", err.Error())
		}
	})
}

func TestEndAllUserSessions(t *testing.T) {
	t.Run("Valid User ID", func(t *testing.T) {
		// Setup
		userID := uuid.New()
		mockRepo := &MockSessionRepository{
			deleteUserSessionsFunc: func(ctx context.Context, id uuid.UUID) error {
				if id == userID {
					return nil
				}
				return errors.New("user not found")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		err := service.EndAllUserSessions(context.Background(), userID)

		// Verify
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Database Error", func(t *testing.T) {
		// Setup
		userID := uuid.New()
		mockRepo := &MockSessionRepository{
			deleteUserSessionsFunc: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("database error")
			},
		}
		service := services.NewSessionService(mockRepo)

		// Execute
		err := service.EndAllUserSessions(context.Background(), userID)

		// Verify
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected error message 'database error', got '%s'", err.Error())
		}
	})
}
