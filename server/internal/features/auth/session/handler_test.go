package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/session"
)

// MockRepository implements session.Repository for testing
type MockRepository struct {
	refreshAccessTokenFunc       func(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)
	endSessionByAccessTokenFunc  func(ctx context.Context, accessToken string) error
	endSessionByRefreshTokenFunc func(ctx context.Context, refreshToken string) error
	endAllUserSessionsFunc       func(ctx context.Context, userID uuid.UUID) error
	getSessionByAccessTokenFunc  func(ctx context.Context, token string) (*models.Session, error)
	getSessionByRefreshTokenFunc func(ctx context.Context, token string) (*models.Session, error)
	createSessionFunc            func(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error)
}

func (m *MockRepository) GetSessionByAccessToken(ctx context.Context, token string) (*models.Session, error) {
	if m.getSessionByAccessTokenFunc != nil {
		return m.getSessionByAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("GetSessionByAccessToken not implemented")
}

func (m *MockRepository) GetSessionByRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	if m.getSessionByRefreshTokenFunc != nil {
		return m.getSessionByRefreshTokenFunc(ctx, token)
	}
	return nil, errors.New("GetSessionByRefreshToken not implemented")
}

func (m *MockRepository) CreateSession(ctx context.Context, userID uuid.UUID, accessExpiry, refreshExpiry time.Duration) (*models.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, userID, accessExpiry, refreshExpiry)
	}
	// Default implementation for handler tests
	return &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   "test_access_token",
		RefreshToken:  "test_refresh_token",
		AccessExpiry:  time.Now().Add(accessExpiry),
		RefreshExpiry: time.Now().Add(refreshExpiry),
	}, nil
}

func (m *MockRepository) RefreshAccessToken(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	if m.refreshAccessTokenFunc != nil {
		return m.refreshAccessTokenFunc(ctx, sessionID)
	}
	return nil, errors.New("RefreshAccessToken not implemented")
}

func (m *MockRepository) DeleteSessionByAccessToken(ctx context.Context, token string) error {
	if m.endSessionByAccessTokenFunc != nil {
		return m.endSessionByAccessTokenFunc(ctx, token)
	}
	return errors.New("EndSessionByAccessToken not implemented")
}

func (m *MockRepository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	if m.endSessionByRefreshTokenFunc != nil {
		return m.endSessionByRefreshTokenFunc(ctx, token)
	}
	return errors.New("EndSessionByRefreshToken not implemented")
}

func (m *MockRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	if m.endAllUserSessionsFunc != nil {
		return m.endAllUserSessionsFunc(ctx, userID)
	}
	return errors.New("DeleteUserSessions not implemented")
}

// Helper functions

// Helper function to create a new test context
func newTestContext(method, path string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// Helper function to add cookies to a request
func addCookies(c echo.Context, cookies ...*http.Cookie) {
	req := c.Request()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

// Helper function to check response status
func checkResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, rec.Code)
	}
}

// Helper to verify cookies are cleared (empty value and negative MaxAge)
func checkCookiesCleared(t *testing.T, rec *httptest.ResponseRecorder, cookieNames ...string) {
	t.Helper()
	cookies := rec.Result().Cookies()

	for _, name := range cookieNames {
		var found bool
		for _, cookie := range cookies {
			if cookie.Name == name {
				found = true
				if cookie.Value != "" {
					t.Errorf("Expected cookie %s value to be empty, got '%s'", name, cookie.Value)
				}
				if cookie.MaxAge >= 0 {
					t.Errorf("Expected cookie %s MaxAge to be negative, got %d", name, cookie.MaxAge)
				}
			}
		}
		if !found {
			t.Errorf("Expected to find cleared cookie %s", name)
		}
	}
}

// Setup creates handler with mock repository for testing
func setupHandler() (*session.Handler, *MockRepository) {
	mockRepo := &MockRepository{}

	// Create service
	service := session.NewService(mockRepo)

	// Create handler
	handler := session.NewHandler(service)

	return handler, mockRepo
}

func TestLogout(t *testing.T) {
	testCases := []struct {
		name              string
		setupCookies      []*http.Cookie
		mockRepoFunc      func(*MockRepository)
		expectedStatus    int
		expectedMessage   string
		cookiesToCheck    []string
		shouldClearCookie bool
	}{
		{
			name: "SuccessfulLogout",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "test_access_token"},
				{Name: "refresh_token", Value: "test_refresh_token"},
			},
			mockRepoFunc: func(mockRepo *MockRepository) {
				mockRepo.endSessionByAccessTokenFunc = func(ctx context.Context, token string) error {
					return nil
				}
				mockRepo.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
					return nil
				}
			},
			expectedStatus:    http.StatusOK,
			expectedMessage:   "Successfully logged out",
			cookiesToCheck:    []string{"access_token", "refresh_token"},
			shouldClearCookie: true,
		},
		{
			name:              "AlreadyLoggedOut",
			setupCookies:      []*http.Cookie{},
			mockRepoFunc:      func(mockRepo *MockRepository) {},
			expectedStatus:    http.StatusOK,
			expectedMessage:   "Already logged out",
			cookiesToCheck:    []string{},
			shouldClearCookie: false,
		},
		{
			name: "AccessTokenErrorButContinue",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "test_access_token"},
				{Name: "refresh_token", Value: "test_refresh_token"},
			},
			mockRepoFunc: func(mockRepo *MockRepository) {
				mockRepo.endSessionByAccessTokenFunc = func(ctx context.Context, token string) error {
					return errors.New("failed to delete session")
				}
				mockRepo.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
					return nil
				}
			},
			expectedStatus:    http.StatusOK,
			expectedMessage:   "Successfully logged out",
			cookiesToCheck:    []string{"access_token", "refresh_token"},
			shouldClearCookie: true,
		},
		{
			name: "RefreshTokenErrorButContinue",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "test_access_token"},
				{Name: "refresh_token", Value: "test_refresh_token"},
			},
			mockRepoFunc: func(mockRepo *MockRepository) {
				mockRepo.endSessionByAccessTokenFunc = func(ctx context.Context, token string) error {
					return nil
				}
				mockRepo.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
					return errors.New("failed to delete session")
				}
			},
			expectedStatus:    http.StatusOK,
			expectedMessage:   "Successfully logged out",
			cookiesToCheck:    []string{"access_token", "refresh_token"},
			shouldClearCookie: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockRepo := setupHandler()

			// Setup request with cookies
			c, rec := newTestContext(http.MethodPost, "/auth/logout", nil)
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Apply mock repo setup
			tc.mockRepoFunc(mockRepo)

			// Execute
			err := handler.LogoutUser(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["message"] != tc.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tc.expectedMessage, response["message"])
			}

			// Check if cookies are cleared
			if tc.shouldClearCookie && len(tc.cookiesToCheck) > 0 {
				checkCookiesCleared(t, rec, tc.cookiesToCheck...)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	testCases := []struct {
		name              string
		setupCookies      []*http.Cookie
		mockRepoFunc      func(*MockRepository)
		expectedStatus    int
		expectedMessage   string
		checkAccessCookie bool
		accessTokenValue  string
	}{
		{
			name: "SuccessfulTokenRefresh",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			mockRepoFunc: func(mockRepo *MockRepository) {
				mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return &models.Session{
						ID:            uuid.New(),
						UserID:        uuid.New(),
						AccessToken:   "old_access_token",
						RefreshToken:  "valid_refresh_token",
						AccessExpiry:  time.Now().Add(-1 * time.Minute), // Expired access token
						RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
					}, nil
				}
				mockRepo.refreshAccessTokenFunc = func(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
					return &models.Session{
						ID:            uuid.New(),
						UserID:        uuid.New(),
						AccessToken:   "new_access_token",
						RefreshToken:  "valid_refresh_token",
						AccessExpiry:  time.Now().Add(15 * time.Minute),
						RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
					}, nil
				}
			},
			expectedStatus:    http.StatusOK,
			expectedMessage:   "Access token refreshed successfully",
			checkAccessCookie: true,
			accessTokenValue:  "new_access_token",
		},
		{
			name:              "NoRefreshTokenProvided",
			setupCookies:      []*http.Cookie{},
			mockRepoFunc:      func(mockRepo *MockRepository) {},
			expectedStatus:    http.StatusUnauthorized,
			expectedMessage:   "No refresh token provided",
			checkAccessCookie: false,
		},
		{
			name: "InvalidRefreshToken",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "invalid_refresh_token"},
			},
			mockRepoFunc: func(mockRepo *MockRepository) {
				mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid refresh token")
				}
			},
			expectedStatus:    http.StatusUnauthorized,
			expectedMessage:   "Invalid refresh token",
			checkAccessCookie: false,
		},
		{
			name: "RefreshAccessTokenError",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			mockRepoFunc: func(mockRepo *MockRepository) {
				mockRepo.getSessionByRefreshTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return &models.Session{
						ID:            uuid.New(),
						UserID:        uuid.New(),
						AccessToken:   "old_access_token",
						RefreshToken:  "valid_refresh_token",
						AccessExpiry:  time.Now().Add(-1 * time.Minute),
						RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
					}, nil
				}
				mockRepo.refreshAccessTokenFunc = func(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
					return nil, errors.New("failed to refresh access token")
				}
			},
			expectedStatus:    http.StatusUnauthorized,
			expectedMessage:   "Invalid refresh token",
			checkAccessCookie: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockRepo := setupHandler()

			// Setup request with cookies
			c, rec := newTestContext(http.MethodPost, "/auth/refresh", nil)
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Apply mock repo setup
			tc.mockRepoFunc(mockRepo)

			// Execute
			err := handler.RefreshToken(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Check for message or error
			if tc.expectedStatus == http.StatusOK {
				if response["message"] != tc.expectedMessage {
					t.Errorf("Expected message '%s', got '%s'", tc.expectedMessage, response["message"])
				}
			} else {
				if response["error"] != tc.expectedMessage {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedMessage, response["error"])
				}
			}

			// Check for new access token cookie
			if tc.checkAccessCookie {
				cookies := rec.Result().Cookies()
				var accessTokenFound bool
				for _, cookie := range cookies {
					if cookie.Name == "access_token" {
						accessTokenFound = true
						if cookie.Value != tc.accessTokenValue {
							t.Errorf("Expected access token '%s', got '%s'", tc.accessTokenValue, cookie.Value)
						}
					}
				}
				if !accessTokenFound {
					t.Error("Access token cookie not set")
				}
			}
		})
	}
}

func TestGetCSRFToken(t *testing.T) {
	t.Run("GetCSRFToken", func(t *testing.T) {
		handler, _ := setupHandler()

		// Setup request
		c, rec := newTestContext(http.MethodGet, "/auth/csrf", nil)

		// Set CSRF token in context
		c.Set("csrf", "test_csrf_token")

		// Execute
		err := handler.GetCSRFToken(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusOK)

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["csrf_token"] != "test_csrf_token" {
			t.Errorf("Expected csrf_token 'test_csrf_token', got '%s'", response["csrf_token"])
		}
	})
}
