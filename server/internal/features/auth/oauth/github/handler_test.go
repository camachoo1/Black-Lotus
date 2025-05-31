package github_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/oauth/github"
)

// MockService mocks the GitHub Service for testing
type MockService struct {
	authenticateFunc func(ctx context.Context, code string) (*models.User, error)
	getAuthURLFunc   func(redirectURI string, state string) string
}

func (m *MockService) Authenticate(ctx context.Context, code string) (*models.User, error) {
	if m.authenticateFunc != nil {
		return m.authenticateFunc(ctx, code)
	}
	return nil, errors.New("not implemented")
}

func (m *MockService) GetAuthURL(redirectURI string, state string) string {
	if m.getAuthURLFunc != nil {
		return m.getAuthURLFunc(redirectURI, state)
	}
	return "https://github.com/auth/url"
}

// MockSessionService mocks the session service
type MockSessionService struct {
	createSessionFunc            func(ctx context.Context, userID uuid.UUID) (*models.Session, error)
	validateAccessTokenFunc      func(ctx context.Context, token string) (*models.Session, error)
	validateRefreshTokenFunc     func(ctx context.Context, token string) (*models.Session, error)
	refreshAccessTokenFunc       func(ctx context.Context, refreshToken string) (*models.Session, error)
	endSessionByAccessTokenFunc  func(ctx context.Context, token string) error
	endSessionByRefreshTokenFunc func(ctx context.Context, token string) error
	endAllUserSessionsFunc       func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSessionService) ValidateAccessToken(ctx context.Context, token string) (*models.Session, error) {
	if m.validateAccessTokenFunc != nil {
		return m.validateAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSessionService) ValidateRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	if m.validateRefreshTokenFunc != nil {
		return m.validateRefreshTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSessionService) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	if m.refreshAccessTokenFunc != nil {
		return m.refreshAccessTokenFunc(ctx, refreshToken)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSessionService) EndSessionByAccessToken(ctx context.Context, token string) error {
	if m.endSessionByAccessTokenFunc != nil {
		return m.endSessionByAccessTokenFunc(ctx, token)
	}
	return errors.New("not implemented")
}

func (m *MockSessionService) EndSessionByRefreshToken(ctx context.Context, token string) error {
	if m.endSessionByRefreshTokenFunc != nil {
		return m.endSessionByRefreshTokenFunc(ctx, token)
	}
	return errors.New("not implemented")
}
func (m *MockSessionService) EndAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	if m.endAllUserSessionsFunc != nil {
		return m.endAllUserSessionsFunc(ctx, userID)
	}
	return errors.New("not implemented")
}

// Helper functions that will be common across tests

// Helper function to create a new test context with the Echo framework
func newTestContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// Helper function to check response status
func checkResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, rec.Code)
	}
}

// Helper function to check JSON response fields
func checkJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedFields map[string]string) {
	t.Helper()
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}

	for key, expectedValue := range expectedFields {
		if value, exists := response[key]; !exists || value != expectedValue {
			t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, value)
		}
	}
}

// Helper to verify token cookies are present and have expected values
func checkTokenCookies(t *testing.T, rec *httptest.ResponseRecorder, expectedValues map[string]string) {
	t.Helper()
	cookies := rec.Result().Cookies()

	var accessTokenFound, refreshTokenFound bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessTokenFound = true
			if expectedValue, exists := expectedValues["access_token"]; exists && cookie.Value != expectedValue {
				t.Errorf("Expected access token %s, got %s", expectedValue, cookie.Value)
			}
		}
		if cookie.Name == "refresh_token" {
			refreshTokenFound = true
			if expectedValue, exists := expectedValues["refresh_token"]; exists && cookie.Value != expectedValue {
				t.Errorf("Expected refresh token %s, got %s", expectedValue, cookie.Value)
			}
		}
	}

	if _, exists := expectedValues["access_token"]; exists && !accessTokenFound {
		t.Error("Access token cookie not set")
	}

	if _, exists := expectedValues["refresh_token"]; exists && !refreshTokenFound {
		t.Error("Refresh token cookie not set")
	}
}

// Helper function to check redirect location
func checkRedirectLocation(t *testing.T, rec *httptest.ResponseRecorder, expectedLocation string) {
	t.Helper()
	location := rec.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("Expected redirect to '%s', got '%s'", expectedLocation, location)
	}
}

// createTestSession creates a test session with the given user ID and tokens
func createTestSession(userID uuid.UUID, accessToken, refreshToken string) *models.Session {
	return &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpiry:  time.Now().Add(15 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	}
}

// TestGetAuthURL tests the GitHub auth URL endpoint
func TestGetAuthURL(t *testing.T) {
	// Test cases for auth URL
	tests := []struct {
		name           string
		path           string
		returnTo       string
		expectedURL    string
		expectedStatus int
	}{
		{
			name:           "With Return Path",
			path:           "/api/auth/github/?returnTo=/dashboard",
			returnTo:       "/dashboard",
			expectedURL:    "https://github.com/auth/url",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With Default Return Path",
			path:           "/api/auth/github",
			returnTo:       "/",
			expectedURL:    "https://github.com/auth/url",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup services
			mockService := &MockService{
				getAuthURLFunc: func(redirectURI, state string) string {
					// Return the expected URL for the test
					return tc.expectedURL
				},
			}
			mockSessionService := &MockSessionService{}

			// Create handler
			handler := github.NewHandler(mockService, mockSessionService)

			// Setup Echo context
			c, rec := newTestContext(http.MethodGet, tc.path)

			// Add return parameter if present in path
			if tc.returnTo != "/" {
				c.SetParamNames("returnTo")
				c.SetParamValues(tc.returnTo)
			}

			// Execute handler
			err := handler.GetAuthURL(c)

			// Verify
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Check response body
			checkJSONResponse(t, rec, map[string]string{
				"url": tc.expectedURL,
			})
		})
	}
}

// TestHandleCallback tests the GitHub callback handler
func TestHandleCallback(t *testing.T) {
	// Set frontend URL for tests
	os.Setenv("FRONTEND_URL", "http://localhost:3000")

	// Test cases for the callback
	testCases := []struct {
		name                string
		path                string
		setupMocks          func(*MockService, *MockSessionService, uuid.UUID)
		expectedStatusCode  int
		expectedRedirectURL string
		expectedTokens      map[string]string
		expectedJSONError   map[string]string
	}{
		{
			name: "Successful Callback",
			path: "/api/auth/github/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockService *MockService, mockSession *MockSessionService, userID uuid.UUID) {
				mockService.authenticateFunc = func(ctx context.Context, code string) (*models.User, error) {
					return &models.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil
				}
				mockSession.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
					return createTestSession(uid, "test-access-token", "test-refresh-token"), nil
				}
			},
			expectedStatusCode:  http.StatusFound,
			expectedRedirectURL: "http://localhost:3000/auth/callback?returnTo=%2Fdashboard",
			expectedTokens: map[string]string{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
			},
		},
		{
			name:                "Missing Code",
			path:                "/api/auth/github/callback?state=%2Fdashboard",
			setupMocks:          func(*MockService, *MockSessionService, uuid.UUID) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRedirectURL: "",
			expectedJSONError: map[string]string{
				"error": "Missing code parameter",
			},
		},
		{
			name: "Authentication Error",
			path: "/api/auth/github/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockService *MockService, mockSession *MockSessionService, userID uuid.UUID) {
				mockService.authenticateFunc = func(ctx context.Context, code string) (*models.User, error) {
					return nil, errors.New("authentication failed")
				}
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRedirectURL: "",
			expectedJSONError: map[string]string{
				"error": "Authentication failed: authentication failed",
			},
		},
		{
			name: "Session Creation Error",
			path: "/api/auth/github/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockService *MockService, mockSession *MockSessionService, userID uuid.UUID) {
				mockService.authenticateFunc = func(ctx context.Context, code string) (*models.User, error) {
					return &models.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil
				}
				mockSession.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
					return nil, errors.New("session creation failed")
				}
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRedirectURL: "",
			expectedJSONError: map[string]string{
				"error": "Failed to create session",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup services
			mockService := &MockService{}
			mockSessionService := &MockSessionService{}

			// Create handler
			handler := github.NewHandler(mockService, mockSessionService)

			// Setup Echo context
			c, rec := newTestContext(http.MethodGet, tc.path)

			// Setup mocks
			userID := uuid.New()
			tc.setupMocks(mockService, mockSessionService, userID)

			// Execute handler
			err := handler.HandleCallback(c)

			// Verify
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatusCode)

			// Check success or error response
			if tc.expectedRedirectURL != "" {
				// For successful cases, check redirect URL and cookies
				checkRedirectLocation(t, rec, tc.expectedRedirectURL)
				checkTokenCookies(t, rec, tc.expectedTokens)
			} else if tc.expectedJSONError != nil {
				// For error cases, check JSON error response
				checkJSONResponse(t, rec, tc.expectedJSONError)
			}
		})
	}
}
