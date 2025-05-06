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

// setupOAuthController creates a controller with mock services for testing
func setupOAuthController() (*controllers.OAuthController, *MockOAuthService, *MockSessionService) {
	mockOAuthService := &MockOAuthService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewOAuthController(mockOAuthService, mockSessionService)

	// Set test environment variables
	os.Setenv("FRONTEND_URL", "http://localhost:3000")

	// Return a cleanup function to restore environment variables
	return controller, mockOAuthService, mockSessionService
}

// Test GetGitHubAuthURL
func TestGetGitHubAuthURL(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, _ := setupOAuthController()

	// Setup Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/github?returnTo=/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mock
	expectedURL := "https://github.com/auth/url"
	mockOAuthService.getAuthorizationURLFunc = func(provider, redirectURI, state string) string {
		if provider != "github" {
			t.Errorf("Expected provider 'github', got '%s'", provider)
		}
		// Check that redirectURI is properly constructed
		if redirectURI != "http://"+req.Host+"/api/auth/github/callback" {
			t.Errorf("Expected redirectURI '%s', got '%s'", "http://"+req.Host+"/api/auth/github/callback", redirectURI)
		}
		// Check that state contains returnTo parameter
		if state != "/dashboard" {
			t.Errorf("Expected state '/dashboard', got '%s'", state)
		}
		return expectedURL
	}

	// Execute
	err := controller.GetGitHubAuthURL(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error parsing response: %v", err)
	}

	if url, exists := response["url"]; !exists || url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, url)
	}
}

// Test GetGitHubAuthURL with default return path
func TestGetGitHubAuthURLWithDefaultReturnPath(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, _ := setupOAuthController()

	// Setup Echo context with no returnTo parameter
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/github", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mock
	mockOAuthService.getAuthorizationURLFunc = func(provider, redirectURI, state string) string {
		// Check that state is the default "/"
		if state != "/" {
			t.Errorf("Expected default state '/', got '%s'", state)
		}
		return "https://github.com/auth/url"
	}

	// Execute
	err := controller.GetGitHubAuthURL(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

// Test HandleGitHubCallback
func TestHandleGitHubCallback(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, mockSessionService := setupOAuthController()

	// Setup Echo context with code parameter
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?code=test-code&state=%2Fdashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mocks
	userID := uuid.New()
	mockUser := &models.User{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
	}

	mockSession := &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   "test-access-token",
		RefreshToken:  "test-refresh-token",
		AccessExpiry:  time.Now().Add(15 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	}

	mockOAuthService.authenticateGitHubFunc = func(ctx context.Context, code string) (*models.User, error) {
		if code != "test-code" {
			t.Errorf("Expected code 'test-code', got '%s'", code)
		}
		return mockUser, nil
	}

	mockSessionService.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
		if uid != userID {
			t.Errorf("Expected user ID %s, got %s", userID, uid)
		}
		return mockSession, nil
	}

	// Execute
	err := controller.HandleGitHubCallback(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check redirection
	if rec.Code != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, rec.Code)
	}

	// Check redirection URL
	expectedRedirectURL := "http://localhost:3000/auth/callback?returnTo=%2Fdashboard"
	location := rec.Header().Get("Location")
	if location != expectedRedirectURL {
		t.Errorf("Expected redirect to '%s', got '%s'", expectedRedirectURL, location)
	}

	// Check cookies
	cookies := rec.Result().Cookies()
	var accessTokenFound, refreshTokenFound bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessTokenFound = true
			if cookie.Value != "test-access-token" {
				t.Errorf("Expected access token 'test-access-token', got '%s'", cookie.Value)
			}
		}
		if cookie.Name == "refresh_token" {
			refreshTokenFound = true
			if cookie.Value != "test-refresh-token" {
				t.Errorf("Expected refresh token 'test-refresh-token', got '%s'", cookie.Value)
			}
		}
	}

	if !accessTokenFound {
		t.Error("Access token cookie not set")
	}
	if !refreshTokenFound {
		t.Error("Refresh token cookie not set")
	}
}

// Test HandleGitHubCallback with missing code
func TestHandleGitHubCallbackWithMissingCode(t *testing.T) {
	// Setup controller with mocks
	controller, _, _ := setupOAuthController()

	// Setup Echo context without code parameter
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state=%2Fdashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := controller.HandleGitHubCallback(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error parsing response: %v", err)
	}

	if errMsg, exists := response["error"]; !exists || errMsg != "Missing code parameter" {
		t.Errorf("Expected error 'Missing code parameter', got '%s'", errMsg)
	}
}

// Test HandleGitHubCallback with authentication error
func TestHandleGitHubCallbackWithAuthError(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, _ := setupOAuthController()

	// Setup Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?code=test-code&state=%2Fdashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mock to return error
	mockOAuthService.authenticateGitHubFunc = func(ctx context.Context, code string) (*models.User, error) {
		return nil, errors.New("authentication failed")
	}

	// Execute
	err := controller.HandleGitHubCallback(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error parsing response: %v", err)
	}

	expectedError := "Authentication failed: authentication failed"
	if errMsg, exists := response["error"]; !exists || errMsg != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errMsg)
	}
}

// Test HandleGitHubCallback with session creation error
func TestHandleGitHubCallbackWithSessionError(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, mockSessionService := setupOAuthController()

	// Setup Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?code=test-code&state=%2Fdashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mocks
	userID := uuid.New()
	mockUser := &models.User{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
	}

	mockOAuthService.authenticateGitHubFunc = func(ctx context.Context, code string) (*models.User, error) {
		return mockUser, nil
	}

	mockSessionService.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
		return nil, errors.New("session creation failed")
	}

	// Execute
	err := controller.HandleGitHubCallback(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error parsing response: %v", err)
	}

	if errMsg, exists := response["error"]; !exists || errMsg != "Failed to create session" {
		t.Errorf("Expected error 'Failed to create session', got '%s'", errMsg)
	}
}

// Test GetGoogleAuthURL
func TestGetGoogleAuthURL(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, _ := setupOAuthController()

	// Setup Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google?returnTo=/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mock
	expectedURL := "https://accounts.google.com/o/oauth2/auth/url"
	mockOAuthService.getAuthorizationURLFunc = func(provider, redirectURI, state string) string {
		if provider != "google" {
			t.Errorf("Expected provider 'google', got '%s'", provider)
		}
		// Check that redirectURI is properly constructed
		if redirectURI != "http://"+req.Host+"/api/auth/google/callback" {
			t.Errorf("Expected redirectURI '%s', got '%s'", "http://"+req.Host+"/api/auth/google/callback", redirectURI)
		}
		// Check that state contains returnTo parameter
		if state != "/dashboard" {
			t.Errorf("Expected state '/dashboard', got '%s'", state)
		}
		return expectedURL
	}

	// Execute
	err := controller.GetGoogleAuthURL(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error parsing response: %v", err)
	}

	if url, exists := response["url"]; !exists || url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, url)
	}
}

// Test HandleGoogleCallback
func TestHandleGoogleCallback(t *testing.T) {
	// Setup controller with mocks
	controller, mockOAuthService, mockSessionService := setupOAuthController()

	// Setup Echo context with code parameter
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=test-code&state=%2Fdashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Configure mocks
	userID := uuid.New()
	mockUser := &models.User{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
	}

	mockSession := &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   "test-access-token",
		RefreshToken:  "test-refresh-token",
		AccessExpiry:  time.Now().Add(15 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	}

	mockOAuthService.authenticateGoogleFunc = func(ctx context.Context, code string, redirectURI string) (*models.User, error) {
		if code != "test-code" {
			t.Errorf("Expected code 'test-code', got '%s'", code)
		}
		expectedRedirectURI := "http://" + req.Host + "/api/auth/google/callback"
		if redirectURI != expectedRedirectURI {
			t.Errorf("Expected redirectURI '%s', got '%s'", expectedRedirectURI, redirectURI)
		}
		return mockUser, nil
	}

	mockSessionService.createSessionFunc = func(ctx context.Context, uid uuid.UUID) (*models.Session, error) {
		if uid != userID {
			t.Errorf("Expected user ID %s, got %s", userID, uid)
		}
		return mockSession, nil
	}

	// Execute
	err := controller.HandleGoogleCallback(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check redirection
	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	// Check redirection URL
	expectedRedirectURL := "http://localhost:3000/auth/callback?returnTo=%2Fdashboard"
	location := rec.Header().Get("Location")
	if location != expectedRedirectURL {
		t.Errorf("Expected redirect to '%s', got '%s'", expectedRedirectURL, location)
	}

	// Check cookies
	cookies := rec.Result().Cookies()
	var accessTokenFound, refreshTokenFound bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessTokenFound = true
			if cookie.Value != "test-access-token" {
				t.Errorf("Expected access token 'test-access-token', got '%s'", cookie.Value)
			}
		}
		if cookie.Name == "refresh_token" {
			refreshTokenFound = true
			if cookie.Value != "test-refresh-token" {
				t.Errorf("Expected refresh token 'test-refresh-token', got '%s'", cookie.Value)
			}
		}
	}

	if !accessTokenFound {
		t.Error("Access token cookie not set")
	}
	if !refreshTokenFound {
		t.Error("Refresh token cookie not set")
	}
}

// Test HandleGoogleCallback with missing code
func TestHandleGoogleCallbackWithMissingCode(t *testing.T) {
	// Setup controller with mocks
	controller, _, _ := setupOAuthController()

	// Setup Echo context without code parameter
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state=%2Fdashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := controller.HandleGoogleCallback(c)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error parsing response: %v", err)
	}

	if errMsg, exists := response["error"]; !exists || errMsg != "Missing code parameter" {
		t.Errorf("Expected error 'Missing code parameter', got '%s'", errMsg)
	}
}
