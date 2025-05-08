// internal/features/auth/register/handler_test.go
package register_test

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

	validation "black-lotus/internal/common/validations"
	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/register"
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

func setupValidator() *validator.Validate {
	v := validator.New()
	validation.RegisterPasswordValidators(v)
	return v
}

// Helper function to create a new test context with the Echo framework
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

// Helper to verify token cookies are present and have expected values
func checkTokenCookies(t *testing.T, rec *httptest.ResponseRecorder, expectAccessToken, expectRefreshToken bool, expectedValues map[string]string) {
	t.Helper()
	cookies := rec.Result().Cookies()

	var accessTokenFound, refreshTokenFound bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessTokenFound = true
			if expectedValues != nil {
				if expectedValue, exists := expectedValues["access_token"]; exists && cookie.Value != expectedValue {
					t.Errorf("Expected access token %s, got %s", expectedValue, cookie.Value)
				}
			}
		}
		if cookie.Name == "refresh_token" {
			refreshTokenFound = true
			if expectedValues != nil {
				if expectedValue, exists := expectedValues["refresh_token"]; exists && cookie.Value != expectedValue {
					t.Errorf("Expected refresh token %s, got %s", expectedValue, cookie.Value)
				}
			}
		}
	}

	if expectAccessToken && !accessTokenFound {
		t.Error("Access token cookie was not set")
	}

	if expectRefreshToken && !refreshTokenFound {
		t.Error("Refresh token cookie was not set")
	}
}

// CreateTestSession creates a test session with the given user ID and tokens
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
func setupHandler() (*register.Handler, *MockRepository, *MockSessionService) {
	mockRepo := NewMockRepository()
	mockSessionService := &MockSessionService{}

	// Create service
	service := register.NewService(mockRepo)

	// Create validator
	validator := setupValidator()

	// Create handler
	handler := register.NewHandler(service, mockSessionService, validator)

	return handler, mockRepo, mockSessionService
}

func TestRegister(t *testing.T) {
	t.Run("SuccessfulRegistration", func(t *testing.T) {
		handler, mockRepo, mockSessionService := setupHandler()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock user repository
		userID := uuid.New()
		createdUser := &models.User{
			ID:            userID,
			Name:          input.Name,
			Email:         input.Email,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockRepo.createUserFunc = func(ctx context.Context, i models.CreateUserInput, hashedPassword *string) (*models.User, error) {
			if i.Email == input.Email && i.Name == input.Name {
				return createdUser, nil
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
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusCreated)

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
		checkTokenCookies(t, rec, true, true, map[string]string{
			"access_token":  "test_access_token",
			"refresh_token": "test_refresh_token",
		})
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create invalid JSON to trigger a binding error
		invalidJSON := []byte(`{"name": "Test User", "email": "test@example.com", "password": }`) // Note the syntax error

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", invalidJSON)

		// Execute
		err := handler.Register(c)
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

		// Create input with validation error (missing required field)
		input := models.CreateUserInput{
			Name:     "", // Empty name should fail validation
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify that we get validation error details
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Validation failed" {
			t.Errorf("Expected 'Validation failed' error, got: %v", response["error"])
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		if len(details) == 0 {
			t.Error("Expected validation error details, got empty map")
		}
	})

	t.Run("PasswordValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with password validation error (no uppercase)
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("password123!"), // Missing uppercase letter
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify that we get correct password validation error
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		passwordError, exists := details["Password"]
		if !exists {
			t.Fatal("Expected Password validation error, but none found")
		}

		if passwordError != "Password must contain at least one uppercase letter" {
			t.Errorf("Expected uppercase letter error, got: %v", passwordError)
		}
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		handler, mockRepo, _ := setupHandler()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "duplicate@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock repository to return duplicate email error
		mockRepo.createUserFunc = func(ctx context.Context, i models.CreateUserInput, hp *string) (*models.User, error) {
			return nil, errors.New("user with this email already exists")
		}

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusConflict)

		// Verify error message
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "user with this email already exists" {
			t.Errorf("Expected duplicate email error, got: %s", response["error"])
		}
	})

	t.Run("OtherRegistrationError", func(t *testing.T) {
		handler, mockRepo, _ := setupHandler()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock repository to return an error
		mockRepo.createUserFunc = func(ctx context.Context, i models.CreateUserInput, hp *string) (*models.User, error) {
			return nil, errors.New("database error")
		}

		// Execute
		err := handler.Register(c)
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

		if response["error"] != "Failed to create user" {
			t.Errorf("Expected 'Failed to create user' error, got: %s", response["error"])
		}
	})

	t.Run("SessionCreationError", func(t *testing.T) {
		handler, mockRepo, mockSessionService := setupHandler()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock user repository
		userID := uuid.New()
		createdUser := &models.User{
			ID:            userID,
			Name:          input.Name,
			Email:         input.Email,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockRepo.createUserFunc = func(ctx context.Context, i models.CreateUserInput, hp *string) (*models.User, error) {
			return createdUser, nil
		}

		// Mock session service to return error
		mockSessionService.createSessionFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return nil, errors.New("failed to create session")
		}

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code - should still be success
		checkResponseStatus(t, rec, http.StatusCreated)

		// Verify response contains user
		var response models.User
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, response.ID)
		}

		// Verify no cookies were set
		checkTokenCookies(t, rec, false, false, nil)
	})

	t.Run("LowercaseLetterValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with password missing lowercase letters
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("PASSWORD123!"), // Missing lowercase letter
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify error details
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		passwordError, exists := details["Password"]
		if !exists {
			t.Fatal("Expected Password validation error, but none found")
		}

		if passwordError != "Password must contain at least one lowercase letter" {
			t.Errorf("Expected lowercase letter error, got: %v", passwordError)
		}
	})

	t.Run("NumberValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with password missing numbers
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password!"), // Missing number
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify error details
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		passwordError, exists := details["Password"]
		if !exists {
			t.Fatal("Expected Password validation error, but none found")
		}

		if passwordError != "Password must contain at least one number" {
			t.Errorf("Expected number error, got: %v", passwordError)
		}
	})

	t.Run("SpecialCharValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with password missing special characters
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123"), // Missing special character
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify error details
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		passwordError, exists := details["Password"]
		if !exists {
			t.Fatal("Expected Password validation error, but none found")
		}

		if passwordError != "Password must contain at least one special character" {
			t.Errorf("Expected special character error, got: %v", passwordError)
		}
	})

	t.Run("EmailValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with invalid email
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "invalid-email", // Not a valid email format
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify error details
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		emailError, exists := details["Email"]
		if !exists {
			t.Fatal("Expected Email validation error, but none found")
		}

		if emailError != "Please enter a valid email address" {
			t.Errorf("Expected email error, got: %v", emailError)
		}
	})

	t.Run("MinLengthValidationError", func(t *testing.T) {
		handler, _, _ := setupHandler()

		// Create input with too short password
		// Assuming there's a min:8 validation on Password
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Pass1!"), // Too short
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Execute
		err := handler.Register(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusBadRequest)

		// Verify error details
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		details, ok := response["details"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected details field to be a map")
		}

		passwordError, exists := details["Password"]
		if !exists {
			t.Fatal("Expected Password validation error, but none found")
		}

		// The message should include "must be at least X characters long"
		if !strings.Contains(passwordError.(string), "must be at least") {
			t.Errorf("Expected min length error, got: %v", passwordError)
		}
	})
}
