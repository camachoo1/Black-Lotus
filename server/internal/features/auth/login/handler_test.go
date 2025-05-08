// internal/features/auth/login/handler_test.go
package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/login"
)

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

// Helper function to create a new test context
func newTestContext(method, path string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// Helper function to check response status
func checkResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, rec.Code)
	}
}

// Helper to verify token cookies are present
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

// CreateTestSession creates a test session
func createTestSession(userID uuid.UUID, accessToken, refreshToken string) *models.Session {
	return &models.Session{
		ID:            uuid.New(),
		UserID:        userID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpiry:  time.Now().Add(15 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:     time.Now(),
	}
}

// Setup creates handler with mock repositories for testing
func setupHandler() (*login.Handler, *MockRepository, *MockSessionService) {
	mockRepo := NewMockRepository()
	mockSessionService := &MockSessionService{}

	// Create service
	service := login.NewService(mockRepo)

	// Create validator
	validator := validator.New()

	// Create handler
	handler := login.NewHandler(service, mockSessionService, validator)

	return handler, mockRepo, mockSessionService
}

func TestLogin(t *testing.T) {
	t.Run("SuccessfulLogin", func(t *testing.T) {
		handler, mockRepo, mockSessionService := setupHandler()

		// Create test input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Mock user repository
		userID := uuid.New()
		loggedInUser := &models.User{
			ID:            userID,
			Name:          "Test User",
			Email:         input.Email,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockRepo.loginUserFunc = func(ctx context.Context, i models.LoginUserInput) (*models.User, error) {
			if i.Email == input.Email && i.Password == input.Password {
				return loggedInUser, nil
			}
			return nil, errors.New("unexpected input")
		}

		// Mock session service
		mockSessionService.createSessionFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			if id == userID {
				return createTestSession(userID, "test_access_token", "test_refresh_token"), nil
			}
			return nil, errors.New("unexpected user ID")
		}

		// Execute
		err := handler.Login(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusOK)

		// Verify response
		var response models.User
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, response.ID)
		}

		// Check if cookies are set
		checkTokenCookies(t, rec, map[string]string{
			"access_token":  "test_access_token",
			"refresh_token": "test_refresh_token",
		})
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create invalid JSON to trigger a binding error
		invalidJSON := []byte(`{"email": "test@example.com", "password": }`) // Note the syntax error

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", invalidJSON)

		// Execute
		err := handler.Login(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify error message
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid request body" {
			t.Errorf("Expected 'Invalid request body' error, got: %s", response["error"])
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with validation error (empty email)
		input := models.LoginUserInput{
			Email:    "", // Empty email should fail validation
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Execute
		err := handler.Login(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify that we get an error response (the exact message will depend on your validator)
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] == "" {
			t.Error("Expected validation error message, got empty string")
		}
	})

	t.Run("LoginError", func(t *testing.T) {
		handler, mockRepo, _ := setupHandler()

		// Create test input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Mock user repository to return error
		mockRepo.loginUserFunc = func(ctx context.Context, i models.LoginUserInput) (*models.User, error) {
			return nil, errors.New("authentication failed")
		}

		// Execute
		err := handler.Login(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusUnauthorized)

		// Verify error message
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		expectedError := "Invalid credentials. Please check your email and password and try again."
		if response["error"] != expectedError {
			t.Errorf("Expected '%s', got: '%s'", expectedError, response["error"])
		}
	})

	t.Run("SessionCreationError", func(t *testing.T) {
		handler, mockRepo, mockSessionService := setupHandler()

		// Create test input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Mock user repository
		userID := uuid.New()
		loggedInUser := &models.User{
			ID:            userID,
			Name:          "Test User",
			Email:         input.Email,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockRepo.loginUserFunc = func(ctx context.Context, i models.LoginUserInput) (*models.User, error) {
			return loggedInUser, nil
		}

		// Mock session service to return error
		mockSessionService.createSessionFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return nil, errors.New("failed to create session")
		}

		// Execute
		err := handler.Login(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusInternalServerError)

		// Verify error message
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !strings.Contains(response["error"], "Failed to create session") {
			t.Errorf("Expected error message to contain 'Failed to create session', got: '%s'", response["error"])
		}
	})
}
