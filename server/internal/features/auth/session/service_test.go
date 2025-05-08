package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/session"
)

// Helper function to setup service for testing
func setupServiceTest() (session.ServiceInterface, *MockRepository) {
	mockRepo := &MockRepository{}
	service := session.NewService(mockRepo)
	return service, mockRepo
}

func TestServiceCreateSession(t *testing.T) {
	testCases := []struct {
		name          string
		userID        uuid.UUID
		mockSetup     func(*testing.T, *MockRepository, uuid.UUID) *models.Session
		expectedError bool
		errorMessage  string
	}{
		{
			name:   "SuccessfulCreateSession",
			userID: uuid.New(),
			mockSetup: func(t *testing.T, repo *MockRepository, userID uuid.UUID) *models.Session {
				expectedSession := &models.Session{
					ID:            uuid.New(),
					UserID:        userID,
					AccessToken:   "new_access_token",
					RefreshToken:  "new_refresh_token",
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
				}

				repo.createSessionFunc = func(ctx context.Context, uid uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
					if uid != userID {
						t.Errorf("Expected userID %s, got %s", userID, uid)
					}
					if accessDuration != session.AccessTokenDuration {
						t.Errorf("Expected access duration %v, got %v", session.AccessTokenDuration, accessDuration)
					}
					if refreshDuration != session.RefreshTokenDuration {
						t.Errorf("Expected refresh duration %v, got %v", session.RefreshTokenDuration, refreshDuration)
					}
					return expectedSession, nil
				}

				return expectedSession
			},
			expectedError: false,
		},
		{
			name:   "CreateSessionError",
			userID: uuid.New(),
			mockSetup: func(t *testing.T, repo *MockRepository, userID uuid.UUID) *models.Session {
				repo.createSessionFunc = func(ctx context.Context, uid uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
					return nil, errors.New("database error")
				}
				return nil
			},
			expectedError: true,
			errorMessage:  "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo := setupServiceTest()
			expectedSession := tc.mockSetup(t, mockRepo, tc.userID)

			// Execute
			result, err := service.CreateSession(context.Background(), tc.userID)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result != expectedSession {
					t.Errorf("Expected session %v, got %v", expectedSession, result)
				}
			}
		})
	}
}

func TestServiceValidateAccessToken(t *testing.T) {
	testCases := []struct {
		name          string
		token         string
		mockSetup     func(*testing.T, *MockRepository, string) *models.Session
		expectedError bool
		errorMessage  string
	}{
		{
			name:  "SuccessfulValidateAccessToken",
			token: "valid_access_token",
			mockSetup: func(t *testing.T, repo *MockRepository, token string) *models.Session {
				expectedSession := &models.Session{
					ID:            uuid.New(),
					UserID:        uuid.New(),
					AccessToken:   token,
					RefreshToken:  "test_refresh_token",
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
				}

				repo.getSessionByAccessTokenFunc = func(ctx context.Context, tkn string) (*models.Session, error) {
					if tkn != token {
						t.Errorf("Expected token '%s', got '%s'", token, tkn)
					}
					return expectedSession, nil
				}

				return expectedSession
			},
			expectedError: false,
		},
		{
			name:  "ValidateAccessTokenError",
			token: "invalid_access_token",
			mockSetup: func(t *testing.T, repo *MockRepository, token string) *models.Session {
				repo.getSessionByAccessTokenFunc = func(ctx context.Context, tkn string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
				return nil
			},
			expectedError: true,
			errorMessage:  "invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo := setupServiceTest()
			expectedSession := tc.mockSetup(t, mockRepo, tc.token)

			// Execute
			result, err := service.ValidateAccessToken(context.Background(), tc.token)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result != expectedSession {
					t.Errorf("Expected session %v, got %v", expectedSession, result)
				}
			}
		})
	}
}

func TestServiceValidateRefreshToken(t *testing.T) {
	testCases := []struct {
		name          string
		token         string
		mockSetup     func(*testing.T, *MockRepository, string) *models.Session
		expectedError bool
		errorMessage  string
	}{
		{
			name:  "SuccessfulValidateRefreshToken",
			token: "valid_refresh_token",
			mockSetup: func(t *testing.T, repo *MockRepository, token string) *models.Session {
				expectedSession := &models.Session{
					ID:            uuid.New(),
					UserID:        uuid.New(),
					AccessToken:   "test_access_token",
					RefreshToken:  token,
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
				}

				repo.getSessionByRefreshTokenFunc = func(ctx context.Context, tkn string) (*models.Session, error) {
					if tkn != token {
						t.Errorf("Expected token '%s', got '%s'", token, tkn)
					}
					return expectedSession, nil
				}

				return expectedSession
			},
			expectedError: false,
		},
		{
			name:  "ValidateRefreshTokenError",
			token: "invalid_refresh_token",
			mockSetup: func(t *testing.T, repo *MockRepository, token string) *models.Session {
				repo.getSessionByRefreshTokenFunc = func(ctx context.Context, tkn string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
				return nil
			},
			expectedError: true,
			errorMessage:  "invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo := setupServiceTest()
			expectedSession := tc.mockSetup(t, mockRepo, tc.token)

			// Execute
			result, err := service.ValidateRefreshToken(context.Background(), tc.token)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result != expectedSession {
					t.Errorf("Expected session %v, got %v", expectedSession, result)
				}
			}
		})
	}
}

func TestServiceEndAllUserSessions(t *testing.T) {
	testCases := []struct {
		name          string
		userID        uuid.UUID
		mockSetup     func(*testing.T, *MockRepository, uuid.UUID)
		expectedError bool
		errorMessage  string
	}{
		{
			name:   "SuccessfulEndAllUserSessions",
			userID: uuid.New(),
			mockSetup: func(t *testing.T, repo *MockRepository, userID uuid.UUID) {
				repo.endAllUserSessionsFunc = func(ctx context.Context, uid uuid.UUID) error {
					if uid != userID {
						t.Errorf("Expected userID %s, got %s", userID, uid)
					}
					return nil
				}
			},
			expectedError: false,
		},
		{
			name:   "EndAllUserSessionsError",
			userID: uuid.New(),
			mockSetup: func(t *testing.T, repo *MockRepository, userID uuid.UUID) {
				repo.endAllUserSessionsFunc = func(ctx context.Context, uid uuid.UUID) error {
					return errors.New("database error")
				}
			},
			expectedError: true,
			errorMessage:  "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			service, mockRepo := setupServiceTest()
			tc.mockSetup(t, mockRepo, tc.userID)

			// Execute
			err := service.EndAllUserSessions(context.Background(), tc.userID)

			// Verify
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorMessage != "" && err.Error() != tc.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
