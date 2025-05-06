package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/domain/auth/services"
	"black-lotus/internal/models"
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

// Implementation of repository methods...
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

func (m *MockSessionRepository) GetSessionByAccessToken(ctx context.Context, token string) (*models.Session, error) {
	if m.getSessionByAccessTokenFunc != nil {
		return m.getSessionByAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("token not found")
}

func (m *MockSessionRepository) GetSessionByRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	if m.getSessionByRefreshTokenFunc != nil {
		return m.getSessionByRefreshTokenFunc(ctx, token)
	}
	return nil, errors.New("token not found")
}

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

func (m *MockSessionRepository) DeleteSessionByAccessToken(ctx context.Context, token string) error {
	if m.deleteSessionByAccessTokenFunc != nil {
		return m.deleteSessionByAccessTokenFunc(ctx, token)
	}
	return nil
}

func (m *MockSessionRepository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	if m.deleteSessionByRefreshTokenFunc != nil {
		return m.deleteSessionByRefreshTokenFunc(ctx, token)
	}
	return nil
}

func (m *MockSessionRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	if m.deleteUserSessionsFunc != nil {
		return m.deleteUserSessionsFunc(ctx, userID)
	}
	return nil
}

// Helper function to create a session for testing
func createTestSession(userID uuid.UUID, accessToken, refreshToken string) *models.Session {
	return &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpiry:  time.Now().Add(time.Hour),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	}
}

func TestSessionService(t *testing.T) {
	t.Run("CreateSession", func(t *testing.T) {
		// Table-driven test for session creation
		testCases := []struct {
			name          string
			setupMock     func(*MockSessionRepository, uuid.UUID)
			expectedError bool
			errorMsg      string
		}{
			{
				name: "SuccessfulCreation",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID) {
					mockRepo.createSessionFunc = func(ctx context.Context, uid uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
						if uid != userID {
							t.Errorf("Expected user ID %s, got %s", userID, uid)
						}
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
					}
				},
				expectedError: false,
			},
			{
				name: "RepositoryError",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID) {
					mockRepo.createSessionFunc = func(ctx context.Context, uid uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error) {
						return nil, errors.New("database error")
					}
				},
				expectedError: true,
				errorMsg:      "database error",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				mockRepo := &MockSessionRepository{}
				userID := uuid.New()

				// Apply mock setup
				tc.setupMock(mockRepo, userID)

				service := services.NewSessionService(mockRepo)

				// Execute
				session, err := service.CreateSession(context.Background(), userID)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if session != nil {
						t.Errorf("Expected nil session, got: %v", session)
					}
					if tc.errorMsg != "" && err.Error() != tc.errorMsg {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
					}
				} else {
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
				}
			})
		}
	})

	t.Run("TokenValidation", func(t *testing.T) {
		// Table-driven test for token validation scenarios
		testCases := []struct {
			name          string
			tokenType     string // "access" or "refresh"
			tokenValue    string
			setupMock     func(*MockSessionRepository, uuid.UUID, string)
			expectedError bool
			errorMsg      string
		}{
			{
				name:       "ValidAccessToken",
				tokenType:  "access",
				tokenValue: "valid_access_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					mockRepo.getSessionByAccessTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						if t == token {
							return createTestSession(userID, token, "valid_refresh_token"), nil
						}
						return nil, errors.New("invalid token")
					}
				},
				expectedError: false,
			},
			{
				name:       "InvalidAccessToken",
				tokenType:  "access",
				tokenValue: "invalid_access_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					mockRepo.getSessionByAccessTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						return nil, errors.New("token not found")
					}
				},
				expectedError: true,
				errorMsg:      "token not found",
			},
			{
				name:       "ValidRefreshToken",
				tokenType:  "refresh",
				tokenValue: "valid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						if t == token {
							return createTestSession(userID, "valid_access_token", token), nil
						}
						return nil, errors.New("invalid token")
					}
				},
				expectedError: false,
			},
			{
				name:       "InvalidRefreshToken",
				tokenType:  "refresh",
				tokenValue: "invalid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						return nil, errors.New("token not found")
					}
				},
				expectedError: true,
				errorMsg:      "token not found",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				mockRepo := &MockSessionRepository{}
				userID := uuid.New()

				// Apply mock setup
				tc.setupMock(mockRepo, userID, tc.tokenValue)

				service := services.NewSessionService(mockRepo)

				// Execute
				var session *models.Session
				var err error
				if tc.tokenType == "access" {
					session, err = service.ValidateAccessToken(context.Background(), tc.tokenValue)
				} else {
					session, err = service.ValidateRefreshToken(context.Background(), tc.tokenValue)
				}

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if session != nil {
						t.Errorf("Expected nil session, got: %v", session)
					}
					if tc.errorMsg != "" && err.Error() != tc.errorMsg {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if session == nil {
						t.Fatal("Expected session to be returned, got nil")
					}
					if session.UserID != userID {
						t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
					}
				}
			})
		}
	})

	t.Run("RefreshAccessToken", func(t *testing.T) {
		// Table-driven test for token refresh scenarios
		testCases := []struct {
			name          string
			refreshToken  string
			setupMock     func(*MockSessionRepository, uuid.UUID, string)
			expectedError bool
			errorMsg      string
		}{
			{
				name:         "SuccessfulRefresh",
				refreshToken: "valid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					sessionID := uuid.New()
					// First mock the lookup of the refresh token
					mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						if t == token {
							return &models.Session{
								ID:            sessionID,
								UserID:        userID,
								AccessToken:   "old_access_token",
								RefreshToken:  token,
								AccessExpiry:  time.Now().Add(time.Hour),
								RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
							}, nil
						}
						return nil, errors.New("invalid token")
					}

					// Then mock the refresh of the access token
					mockRepo.refreshAccessTokenFunc = func(ctx context.Context, sID uuid.UUID) (*models.Session, error) {
						if sID == sessionID {
							return &models.Session{
								ID:            sessionID,
								UserID:        userID,
								AccessToken:   "new_access_token",
								RefreshToken:  token,
								AccessExpiry:  time.Now().Add(time.Hour),
								RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
							}, nil
						}
						return nil, errors.New("session not found")
					}
				},
				expectedError: false,
			},
			{
				name:         "InvalidRefreshToken",
				refreshToken: "invalid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						return nil, errors.New("token not found")
					}
				},
				expectedError: true,
				errorMsg:      "token not found",
			},
			{
				name:         "RefreshError",
				refreshToken: "valid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, userID uuid.UUID, token string) {
					sessionID := uuid.New()
					// First mock the lookup of the refresh token
					mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, t string) (*models.Session, error) {
						return &models.Session{
							ID:           sessionID,
							UserID:       userID,
							AccessToken:  "old_access_token",
							RefreshToken: token,
						}, nil
					}

					// Then mock a failure in refreshing the token
					mockRepo.refreshAccessTokenFunc = func(ctx context.Context, sID uuid.UUID) (*models.Session, error) {
						return nil, errors.New("database error")
					}
				},
				expectedError: true,
				errorMsg:      "database error",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				mockRepo := &MockSessionRepository{}
				userID := uuid.New()

				// Apply mock setup
				tc.setupMock(mockRepo, userID, tc.refreshToken)

				service := services.NewSessionService(mockRepo)

				// Execute
				session, err := service.RefreshAccessToken(context.Background(), tc.refreshToken)

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if session != nil {
						t.Errorf("Expected nil session, got: %v", session)
					}
					if tc.errorMsg != "" && err.Error() != tc.errorMsg {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
					if session == nil {
						t.Fatal("Expected session to be returned, got nil")
					}
					if session.AccessToken != "new_access_token" {
						t.Errorf("Expected new access token 'new_access_token', got '%s'", session.AccessToken)
					}
				}
			})
		}
	})

	t.Run("EndSession", func(t *testing.T) {
		// Table-driven test for session ending operations
		testCases := []struct {
			name          string
			methodType    string    // "access", "refresh", or "all"
			token         string    // for access or refresh token methods
			userID        uuid.UUID // for end all sessions method
			setupMock     func(*MockSessionRepository, string, uuid.UUID)
			expectedError bool
			errorMsg      string
		}{
			{
				name:       "EndSessionByValidAccessToken",
				methodType: "access",
				token:      "valid_access_token",
				setupMock: func(mockRepo *MockSessionRepository, token string, _ uuid.UUID) {
					mockRepo.deleteSessionByAccessTokenFunc = func(ctx context.Context, t string) error {
						if t == token {
							return nil
						}
						return errors.New("token not found")
					}
				},
				expectedError: false,
			},
			{
				name:       "EndSessionByInvalidAccessToken",
				methodType: "access",
				token:      "invalid_access_token",
				setupMock: func(mockRepo *MockSessionRepository, token string, _ uuid.UUID) {
					mockRepo.deleteSessionByAccessTokenFunc = func(ctx context.Context, t string) error {
						return errors.New("token not found")
					}
				},
				expectedError: true,
				errorMsg:      "token not found",
			},
			{
				name:       "EndSessionByValidRefreshToken",
				methodType: "refresh",
				token:      "valid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, token string, _ uuid.UUID) {
					mockRepo.deleteSessionByRefreshTokenFunc = func(ctx context.Context, t string) error {
						if t == token {
							return nil
						}
						return errors.New("token not found")
					}
				},
				expectedError: false,
			},
			{
				name:       "EndSessionByInvalidRefreshToken",
				methodType: "refresh",
				token:      "invalid_refresh_token",
				setupMock: func(mockRepo *MockSessionRepository, token string, _ uuid.UUID) {
					mockRepo.deleteSessionByRefreshTokenFunc = func(ctx context.Context, t string) error {
						return errors.New("token not found")
					}
				},
				expectedError: true,
				errorMsg:      "token not found",
			},
			{
				name:       "EndAllUserSessions",
				methodType: "all",
				userID:     uuid.New(),
				setupMock: func(mockRepo *MockSessionRepository, _ string, userID uuid.UUID) {
					mockRepo.deleteUserSessionsFunc = func(ctx context.Context, uid uuid.UUID) error {
						if uid == userID {
							return nil
						}
						return errors.New("user not found")
					}
				},
				expectedError: false,
			},
			{
				name:       "EndAllUserSessionsDBError",
				methodType: "all",
				userID:     uuid.New(),
				setupMock: func(mockRepo *MockSessionRepository, _ string, userID uuid.UUID) {
					mockRepo.deleteUserSessionsFunc = func(ctx context.Context, uid uuid.UUID) error {
						return errors.New("database error")
					}
				},
				expectedError: true,
				errorMsg:      "database error",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				mockRepo := &MockSessionRepository{}

				// Apply mock setup
				tc.setupMock(mockRepo, tc.token, tc.userID)

				service := services.NewSessionService(mockRepo)

				// Execute
				var err error
				switch tc.methodType {
				case "access":
					err = service.EndSessionByAccessToken(context.Background(), tc.token)
				case "refresh":
					err = service.EndSessionByRefreshToken(context.Background(), tc.token)
				case "all":
					err = service.EndAllUserSessions(context.Background(), tc.userID)
				}

				// Verify
				if tc.expectedError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					if tc.errorMsg != "" && err.Error() != tc.errorMsg {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got: %v", err)
					}
				}
			})
		}
	})
}
