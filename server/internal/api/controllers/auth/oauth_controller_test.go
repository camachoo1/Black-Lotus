package controllers_test

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

	controllers "black-lotus/internal/api/controllers/auth"
	"black-lotus/internal/models"
)

// MockOAuthService mocks the OAuthService for testing
type MockOAuthService struct {
	getAuthorizationURLFunc func(provider string, redirectURI string, state string) string
	authenticateGitHubFunc  func(ctx context.Context, code string) (*models.User, error)
	authenticateGoogleFunc  func(ctx context.Context, code string, redirectURI string) (*models.User, error)
}

func (m *MockOAuthService) GetAuthorizationURL(provider string, redirectURI string, state string) string {
	if m.getAuthorizationURLFunc != nil {
		return m.getAuthorizationURLFunc(provider, redirectURI, state)
	}
	return "https://mock-oauth-url.com"
}

func (m *MockOAuthService) AuthenticateGitHub(ctx context.Context, code string) (*models.User, error) {
	if m.authenticateGitHubFunc != nil {
		return m.authenticateGitHubFunc(ctx, code)
	}
	return nil, errors.New("not implemented")
}

func (m *MockOAuthService) AuthenticateGoogle(ctx context.Context, code string, redirectURI string) (*models.User, error) {
	if m.authenticateGoogleFunc != nil {
		return m.authenticateGoogleFunc(ctx, code, redirectURI)
	}
	return nil, errors.New("not implemented")
}

// Helper function to create a new test context with the Echo framework
func newOAuthTestContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// Helper function to check response status
func checkOAuthResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
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
func checkOAuthTokenCookies(t *testing.T, rec *httptest.ResponseRecorder, expectedValues map[string]string) {
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

// createTestOAuthSession creates a test session with the given user ID and tokens
func createTestOAuthSession(userID uuid.UUID, accessToken, refreshToken string) *models.Session {
	return &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpiry:  time.Now().Add(15 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	}
}

// setupOAuthController creates a controller with mock services for testing
func setupOAuthController() (*controllers.OAuthController, *MockOAuthService, *MockSessionService) {
	mockOAuthService := &MockOAuthService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewOAuthController(mockOAuthService, mockSessionService)

	// Set test environment variables
	os.Setenv("FRONTEND_URL", "http://localhost:3000")

	return controller, mockOAuthService, mockSessionService
}

// TestGetProviderAuthURL tests both GitHub and Google auth URL endpoints
func TestGetProviderAuthURL(t *testing.T) {
	// Table-driven test for auth URL endpoints
	tests := []struct {
		name           string
		provider       string
		path           string
		returnTo       string
		expectedURL    string
		expectedStatus int
	}{
		{
			name:           "GitHub with Return Path",
			provider:       "github",
			path:           "/api/auth/github?returnTo=/dashboard",
			returnTo:       "/dashboard",
			expectedURL:    "https://github.com/auth/url",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GitHub with Default Return Path",
			provider:       "github",
			path:           "/api/auth/github",
			returnTo:       "/",
			expectedURL:    "https://github.com/auth/url",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Google with Return Path",
			provider:       "google",
			path:           "/api/auth/google?returnTo=/dashboard",
			returnTo:       "/dashboard",
			expectedURL:    "https://accounts.google.com/o/oauth2/auth/url",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Google with Default Return Path",
			provider:       "google",
			path:           "/api/auth/google",
			returnTo:       "/",
			expectedURL:    "https://accounts.google.com/o/oauth2/auth/url",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller, mockOAuthService, _ := setupOAuthController()

			// Setup Echo context
			c, rec := newOAuthTestContext(http.MethodGet, tc.path)

			// Configure mock
			mockOAuthService.getAuthorizationURLFunc = func(provider, redirectURI, state string) string {
				if provider != tc.provider {
					t.Errorf("Expected provider '%s', got '%s'", tc.provider, provider)
				}
				// Check redirectURI is properly constructed
				expectedRedirectURI := "http://" + c.Request().Host + "/api/auth/" + tc.provider + "/callback"
				if redirectURI != expectedRedirectURI {
					t.Errorf("Expected redirectURI '%s', got '%s'", expectedRedirectURI, redirectURI)
				}
				// Check state contains returnTo parameter
				if state != tc.returnTo {
					t.Errorf("Expected state '%s', got '%s'", tc.returnTo, state)
				}
				return tc.expectedURL
			}

			// Execute the appropriate handler based on provider
			var err error
			if tc.provider == "github" {
				err = controller.GetGitHubAuthURL(c)
			} else {
				err = controller.GetGoogleAuthURL(c)
			}

			// Verify
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkOAuthResponseStatus(t, rec, tc.expectedStatus)

			// Check response body
			checkJSONResponse(t, rec, map[string]string{
				"url": tc.expectedURL,
			})
		})
	}
}

// TestHandleOAuthCallback tests both GitHub and Google callback handlers
func TestHandleOAuthCallback(t *testing.T) {
	// Common test cases for both providers
	testCases := []struct {
		name                string
		provider            string
		path                string
		setupMocks          func(*MockOAuthService, *MockSessionService, uuid.UUID)
		expectedStatusCode  int
		expectedRedirectURL string
		expectedTokens      map[string]string
		expectedJSONError   map[string]string
	}{
		{
			name:     "Successful GitHub Callback",
			provider: "github",
			path:     "/api/auth/github/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockOAuth *MockOAuthService, mockSession *MockSessionService, userID uuid.UUID) {
				mockOAuth.authenticateGitHubFunc = func(ctx context.Context, code string) (*models.User, error) {
					return &models.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil
				}
				mockSession.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
					return createTestOAuthSession(uid, "test-access-token", "test-refresh-token"), nil
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
			name:     "Successful Google Callback",
			provider: "google",
			path:     "/api/auth/google/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockOAuth *MockOAuthService, mockSession *MockSessionService, userID uuid.UUID) {
				mockOAuth.authenticateGoogleFunc = func(ctx context.Context, code string, redirectURI string) (*models.User, error) {
					return &models.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil
				}
				mockSession.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
					return createTestOAuthSession(uid, "test-access-token", "test-refresh-token"), nil
				}
			},
			expectedStatusCode:  http.StatusTemporaryRedirect,
			expectedRedirectURL: "http://localhost:3000/auth/callback?returnTo=%2Fdashboard",
			expectedTokens: map[string]string{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
			},
		},
		{
			name:                "GitHub Missing Code",
			provider:            "github",
			path:                "/api/auth/github/callback?state=%2Fdashboard",
			setupMocks:          func(*MockOAuthService, *MockSessionService, uuid.UUID) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRedirectURL: "",
			expectedJSONError: map[string]string{
				"error": "Missing code parameter",
			},
		},
		{
			name:                "Google Missing Code",
			provider:            "google",
			path:                "/api/auth/google/callback?state=%2Fdashboard",
			setupMocks:          func(*MockOAuthService, *MockSessionService, uuid.UUID) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRedirectURL: "",
			expectedJSONError: map[string]string{
				"error": "Missing code parameter",
			},
		},
		{
			name:     "GitHub Authentication Error",
			provider: "github",
			path:     "/api/auth/github/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockOAuth *MockOAuthService, mockSession *MockSessionService, userID uuid.UUID) {
				mockOAuth.authenticateGitHubFunc = func(ctx context.Context, code string) (*models.User, error) {
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
			name:     "Google Authentication Error",
			provider: "google",
			path:     "/api/auth/google/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockOAuth *MockOAuthService, mockSession *MockSessionService, userID uuid.UUID) {
				mockOAuth.authenticateGoogleFunc = func(ctx context.Context, code string, redirectURI string) (*models.User, error) {
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
			name:     "GitHub Session Creation Error",
			provider: "github",
			path:     "/api/auth/github/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockOAuth *MockOAuthService, mockSession *MockSessionService, userID uuid.UUID) {
				mockOAuth.authenticateGitHubFunc = func(ctx context.Context, code string) (*models.User, error) {
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
		{
			name:     "Google Session Creation Error",
			provider: "google",
			path:     "/api/auth/google/callback?code=test-code&state=%2Fdashboard",
			setupMocks: func(mockOAuth *MockOAuthService, mockSession *MockSessionService, userID uuid.UUID) {
				mockOAuth.authenticateGoogleFunc = func(ctx context.Context, code string, redirectURI string) (*models.User, error) {
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
			controller, mockOAuthService, mockSessionService := setupOAuthController()

			// Setup Echo context
			c, rec := newOAuthTestContext(http.MethodGet, tc.path)

			// Setup mocks
			userID := uuid.New()
			tc.setupMocks(mockOAuthService, mockSessionService, userID)

			// Execute the appropriate handler based on provider
			var err error
			if tc.provider == "github" {
				err = controller.HandleGitHubCallback(c)
			} else {
				err = controller.HandleGoogleCallback(c)
			}

			// Verify
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkOAuthResponseStatus(t, rec, tc.expectedStatusCode)

			// Check success or error response
			if tc.expectedRedirectURL != "" {
				// For successful cases, check redirect URL and cookies
				checkRedirectLocation(t, rec, tc.expectedRedirectURL)
				checkOAuthTokenCookies(t, rec, tc.expectedTokens)
			} else if tc.expectedJSONError != nil {
				// For error cases, check JSON error response
				checkJSONResponse(t, rec, tc.expectedJSONError)
			}
		})
	}
}
