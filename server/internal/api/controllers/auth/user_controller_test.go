package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	controllers "black-lotus/internal/api/controllers/auth"
	"black-lotus/internal/domain/auth/services"
	"black-lotus/internal/models"
)

// MockUserService implements services.UserServiceInterface
type MockUserService struct {
	createUserFunc       func(ctx context.Context, input models.CreateUserInput) (*models.User, error)
	loginUserFunc        func(ctx context.Context, input models.LoginUserInput) (*models.User, error)
	getUserByIDFunc      func(ctx context.Context, userID uuid.UUID) (*models.User, error)
	getUserWithTripsFunc func(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error)
}

// Ensure MockUserService implements services.UserServiceInterface
var _ services.UserServiceInterface = &MockUserService{}

func (m *MockUserService) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, input)
	}
	return nil, errors.New("CreateUser not implemented")
}

func (m *MockUserService) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	if m.loginUserFunc != nil {
		return m.loginUserFunc(ctx, input)
	}
	return nil, errors.New("LoginUser not implemented")
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented")
}

func (m *MockUserService) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error) {
	if m.getUserWithTripsFunc != nil {
		return m.getUserWithTripsFunc(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetUserWithTrips not implemented")
}

// MockSessionService implements services.SessionServiceInterface
type MockSessionService struct {
	createSessionFunc            func(ctx context.Context, userID uuid.UUID) (*models.Session, error)
	validateAccessTokenFunc      func(ctx context.Context, token string) (*models.Session, error)
	validateRefreshTokenFunc     func(ctx context.Context, token string) (*models.Session, error)
	refreshAccessTokenFunc       func(ctx context.Context, refreshToken string) (*models.Session, error)
	endSessionByAccessTokenFunc  func(ctx context.Context, accessToken string) error
	endSessionByRefreshTokenFunc func(ctx context.Context, refreshToken string) error
	endAllUserSessionsFunc       func(ctx context.Context, userID uuid.UUID) error
}

// Ensure MockSessionService implements services.SessionServiceInterface
var _ services.SessionServiceInterface = &MockSessionService{}

func (m *MockSessionService) CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, userID)
	}
	return nil, errors.New("CreateSession not implemented")
}

func (m *MockSessionService) ValidateAccessToken(ctx context.Context, token string) (*models.Session, error) {
	if m.validateAccessTokenFunc != nil {
		return m.validateAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("ValidateAccessToken not implemented")
}

func (m *MockSessionService) ValidateRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	if m.validateAccessTokenFunc != nil {
		return m.validateAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("ValidateAccessToken not implemented")
}

func (m *MockSessionService) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	if m.refreshAccessTokenFunc != nil {
		return m.refreshAccessTokenFunc(ctx, refreshToken)
	}
	return nil, errors.New("RefreshAccessToken not implemented")
}

func (m *MockSessionService) EndSessionByAccessToken(ctx context.Context, accessToken string) error {
	if m.endSessionByAccessTokenFunc != nil {
		return m.endSessionByAccessTokenFunc(ctx, accessToken)
	}
	return errors.New("EndSessionByAccessToken not implemented")
}

func (m *MockSessionService) EndSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	if m.endSessionByRefreshTokenFunc != nil {
		return m.endSessionByRefreshTokenFunc(ctx, refreshToken)
	}
	return errors.New("EndSessionByRefreshToken not implemented")
}

func (m *MockSessionService) EndAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	if m.endAllUserSessionsFunc != nil {
		return m.endAllUserSessionsFunc(ctx, userID)
	}
	return errors.New("EndSessionByRefreshToken not implemented")
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return s != "" && strings.Contains(s, substr)
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create a new test context with the Echo framework
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

// Helper function to check if response status matches expected
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

	if expectAccessToken && !accessTokenFound {
		t.Error("Access token cookie was not set")
	}

	if expectRefreshToken && !refreshTokenFound {
		t.Error("Refresh token cookie was not set")
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

// Setup creates controller with mock services for testing
func setupController() (*controllers.UserController, *MockUserService, *MockSessionService) {
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)
	return controller, mockUserService, mockSessionService
}

func TestRegisterUser(t *testing.T) {
	t.Run("SuccessfulRegistration", func(t *testing.T) {
		controller, mockUserService, mockSessionService := setupController()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock user service
		userID := uuid.New()
		createdUser := &models.User{
			ID:            userID,
			Name:          input.Name,
			Email:         input.Email,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockUserService.createUserFunc = func(ctx context.Context, i models.CreateUserInput) (*models.User, error) {
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
		err := controller.RegisterUser(c)
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

	t.Run("UserAlreadyExists", func(t *testing.T) {
		controller, mockUserService, _ := setupController()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Existing User",
			Email:    "existing@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock user service
		mockUserService.createUserFunc = func(ctx context.Context, i models.CreateUserInput) (*models.User, error) {
			return nil, errors.New("user with this email already exists")
		}

		// Execute
		err := controller.RegisterUser(c)
		if err != nil {
			t.Errorf("Expected no error from handler, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusConflict)

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "user with this email already exists" {
			t.Errorf("Expected error 'user with this email already exists', got '%s'", response["error"])
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Table-driven test for validation errors
		testCases := []struct {
			name          string
			input         models.CreateUserInput
			expectedField string
			expectedMsg   string
		}{
			{
				name: "MissingPassword",
				input: models.CreateUserInput{
					Name:  "Test User",
					Email: "test@example.com",
					// Password is nil
				},
				expectedField: "password",
				expectedMsg:   "password is required",
			},
			{
				name: "InvalidEmail",
				input: models.CreateUserInput{
					Name:     "Test User",
					Email:    "not-an-email",
					Password: stringPtr("Password123!"),
				},
				expectedField: "email",
				expectedMsg:   "Please enter a valid email address",
			},
			{
				name: "PasswordTooShort",
				input: models.CreateUserInput{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: stringPtr("Pa1!"), // Too short
				},
				expectedField: "password",
				expectedMsg:   "password must be at least 6 characters long",
			},
			{
				name: "PasswordWithoutUppercase",
				input: models.CreateUserInput{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: stringPtr("password123!"),
				},
				expectedField: "password",
				expectedMsg:   "Password must contain at least one uppercase letter",
			},
			{
				name: "PasswordWithoutLowercase",
				input: models.CreateUserInput{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: stringPtr("PASSWORD123!"),
				},
				expectedField: "password",
				expectedMsg:   "Password must contain at least one lowercase letter",
			},
			{
				name: "PasswordWithoutNumber",
				input: models.CreateUserInput{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: stringPtr("Password!"),
				},
				expectedField: "password",
				expectedMsg:   "Password must contain at least one number",
			},
			{
				name: "PasswordWithoutSpecialChar",
				input: models.CreateUserInput{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: stringPtr("Password123"),
				},
				expectedField: "password",
				expectedMsg:   "Password must contain at least one special character",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				controller, _, _ := setupController()

				// Setup request with test case input
				inputJSON, _ := json.Marshal(tc.input)
				c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

				// Execute
				err := controller.RegisterUser(c)
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Check status code
				checkResponseStatus(t, rec, http.StatusBadRequest)

				// Verify response contains validation error
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				errorMsg, ok := response["error"].(string)
				if !ok || errorMsg != "Validation failed" {
					t.Errorf("Expected error 'Validation failed', got '%v'", response["error"])
				}

				details, ok := response["details"].(map[string]interface{})
				if !ok || details == nil {
					t.Errorf("Expected validation details, got nil or wrong type: %v", response)
					return
				}

				// Check specific error for the field
				fieldError, exists := details[tc.expectedField]
				if !exists {
					t.Errorf("Expected validation error for field '%s', but none found in: %v",
						tc.expectedField, details)
					return
				}

				errorString, ok := fieldError.(string)
				if !ok {
					t.Errorf("Expected error message to be string, got %T: %v", fieldError, fieldError)
					return
				}

				if errorString != tc.expectedMsg {
					t.Errorf("Error message doesn't match. Got: '%s', Expected: '%s'",
						errorString, tc.expectedMsg)
				}
			})
		}
	})

	t.Run("SessionCreationFailure", func(t *testing.T) {
		controller, mockUserService, mockSessionService := setupController()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock user service success
		userID := uuid.New()
		createdUser := &models.User{
			ID:            userID,
			Name:          input.Name,
			Email:         input.Email,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockUserService.createUserFunc = func(ctx context.Context, i models.CreateUserInput) (*models.User, error) {
			return createdUser, nil
		}

		// Mock session service failure
		mockSessionService.createSessionFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return nil, errors.New("session creation failed")
		}

		// Execute
		err := controller.RegisterUser(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should still be created despite session failure
		checkResponseStatus(t, rec, http.StatusCreated)

		// Verify user is returned
		var response models.User
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, response.ID)
		}

		// Verify no cookies are set
		cookies := rec.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
				t.Errorf("Expected no cookie %s to be set after session creation failure", cookie.Name)
			}
		}
	})

	t.Run("GeneralFailure", func(t *testing.T) {
		controller, mockUserService, _ := setupController()

		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/register", inputJSON)

		// Mock user service to return a general error
		mockUserService.createUserFunc = func(ctx context.Context, i models.CreateUserInput) (*models.User, error) {
			return nil, errors.New("database connection failed")
		}

		// Execute
		err := controller.RegisterUser(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusInternalServerError)

		// Verify error response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Failed to create user" {
			t.Errorf("Expected error 'Failed to create user', got '%s'", response["error"])
		}
	})
}

func TestLoginUser(t *testing.T) {
	t.Run("SuccessfulLogin", func(t *testing.T) {
		controller, mockUserService, mockSessionService := setupController()

		// Create test input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Mock user service
		userID := uuid.New()
		loggedInUser := &models.User{
			ID:            userID,
			Name:          "Test User",
			Email:         input.Email,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockUserService.loginUserFunc = func(ctx context.Context, i models.LoginUserInput) (*models.User, error) {
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
		err := controller.LoginUser(c)
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
		checkTokenCookies(t, rec, true, true, map[string]string{
			"access_token":  "test_access_token",
			"refresh_token": "test_refresh_token",
		})
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		controller, mockUserService, _ := setupController()

		// Create test input
		input := models.LoginUserInput{
			Email:    "wrong@example.com",
			Password: "WrongPassword",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Mock user service
		mockUserService.loginUserFunc = func(ctx context.Context, i models.LoginUserInput) (*models.User, error) {
			return nil, errors.New("invalid credentials")
		}

		// Execute
		err := controller.LoginUser(c)
		if err != nil {
			t.Errorf("Expected no error from handler, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusUnauthorized)

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("SessionCreationFailure", func(t *testing.T) {
		controller, mockUserService, mockSessionService := setupController()

		// Create test input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		c, rec := newTestContext(http.MethodPost, "/auth/login", inputJSON)

		// Mock user service success
		userID := uuid.New()
		loggedInUser := &models.User{
			ID:            userID,
			Name:          "Test User",
			Email:         input.Email,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockUserService.loginUserFunc = func(ctx context.Context, i models.LoginUserInput) (*models.User, error) {
			return loggedInUser, nil
		}

		// Mock session service failure
		mockSessionService.createSessionFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return nil, errors.New("session creation failed")
		}

		// Execute
		err := controller.LoginUser(c)
		if err != nil {
			t.Errorf("Expected no error from handler, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusInternalServerError)

		// Verify error response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !containsSubstring(response["error"], "Failed to create session") {
			t.Errorf("Expected error message containing 'Failed to create session', got '%s'", response["error"])
		}
	})
}

func TestLogoutUser(t *testing.T) {
	t.Run("SuccessfulLogout", func(t *testing.T) {
		controller, _, mockSessionService := setupController()

		// Setup request with cookies
		c, rec := newTestContext(http.MethodPost, "/auth/logout", nil)
		addCookies(c,
			&http.Cookie{Name: "access_token", Value: "test_access_token"},
			&http.Cookie{Name: "refresh_token", Value: "test_refresh_token"},
		)

		// Mock session service
		mockSessionService.endSessionByAccessTokenFunc = func(ctx context.Context, token string) error {
			if token == "test_access_token" {
				return nil
			}
			return errors.New("unexpected token")
		}

		mockSessionService.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
			if token == "test_refresh_token" {
				return nil
			}
			return errors.New("unexpected token")
		}

		// Execute
		err := controller.LogoutUser(c)
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

		if response["message"] != "Successfully logged out" {
			t.Errorf("Expected message 'Successfully logged out', got '%s'", response["message"])
		}

		// Check if cookies are cleared
		checkCookiesCleared(t, rec, "access_token", "refresh_token")
	})

	t.Run("AlreadyLoggedOut", func(t *testing.T) {
		controller, _, _ := setupController()

		// Setup request without cookies
		c, rec := newTestContext(http.MethodPost, "/auth/logout", nil)

		// Execute
		err := controller.LogoutUser(c)
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

		if response["message"] != "Already logged out" {
			t.Errorf("Expected message 'Already logged out', got '%s'", response["message"])
		}
	})

	t.Run("PartialFailure", func(t *testing.T) {
		controller, _, mockSessionService := setupController()

		// Setup request with cookies
		c, rec := newTestContext(http.MethodPost, "/auth/logout", nil)
		addCookies(c,
			&http.Cookie{Name: "access_token", Value: "test_access_token"},
			&http.Cookie{Name: "refresh_token", Value: "test_refresh_token"},
		)

		// Mock endSessionByAccessToken to fail
		mockSessionService.endSessionByAccessTokenFunc = func(ctx context.Context, token string) error {
			return errors.New("access token deletion failed")
		}

		// Mock endSessionByRefreshToken to succeed
		mockSessionService.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
			return nil
		}

		// Execute (this uses the controller and creates response in rec)
		if err := controller.LogoutUser(c); err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code (this uses rec)
		checkResponseStatus(t, rec, http.StatusOK)

		// Verify response (this uses rec)
		var response map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Successfully logged out" {
			t.Errorf("Expected message 'Successfully logged out', got '%s'", response["message"])
		}

		// Check if cookies are cleared (this uses rec)
		checkCookiesCleared(t, rec, "access_token", "refresh_token")
	})

	t.Run("RefreshTokenFailure", func(t *testing.T) {
		controller, _, mockSessionService := setupController()

		// Setup request with only refresh token
		c, rec := newTestContext(http.MethodPost, "/auth/logout", nil)
		addCookies(c, &http.Cookie{Name: "refresh_token", Value: "test_refresh_token"})

		// Mock session service for refresh token failure
		mockSessionService.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
			return errors.New("database error when ending refresh token session")
		}

		// Execute
		err := controller.LogoutUser(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code - should still succeed
		checkResponseStatus(t, rec, http.StatusOK)

		// Verify response indicates success
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Successfully logged out" {
			t.Errorf("Expected message 'Successfully logged out', got '%s'", response["message"])
		}

		// Check if cookies are cleared
		checkCookiesCleared(t, rec, "refresh_token")
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("SuccessfulTokenRefresh", func(t *testing.T) {
		controller, _, mockSessionService := setupController()

		// Setup request with refresh token cookie
		c, rec := newTestContext(http.MethodPost, "/auth/refresh", nil)
		addCookies(c, &http.Cookie{Name: "refresh_token", Value: "valid_refresh_token"})

		// Mock service response
		mockSessionService.refreshAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			if token == "valid_refresh_token" {
				return &models.Session{
					ID:            uuid.New(),
					UserID:        uuid.New(),
					AccessToken:   "new_access_token",
					RefreshToken:  "valid_refresh_token",
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
				}, nil
			}
			return nil, errors.New("invalid refresh token")
		}

		// Execute
		err := controller.RefreshToken(c)
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

		if response["message"] != "Access token refreshed successfully" {
			t.Errorf("Expected message 'Access token refreshed successfully', got '%s'", response["message"])
		}

		// Check if new access token cookie is set
		checkTokenCookies(t, rec, true, false, map[string]string{
			"access_token": "new_access_token",
		})
	})

	t.Run("NoRefreshToken", func(t *testing.T) {
		controller, _, _ := setupController()

		// Setup request without refresh token cookie
		c, rec := newTestContext(http.MethodPost, "/auth/refresh", nil)

		// Execute
		err := controller.RefreshToken(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusUnauthorized)

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "No refresh token provided" {
			t.Errorf("Expected error 'No refresh token provided', got '%s'", response["error"])
		}
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		controller, _, mockSessionService := setupController()

		// Setup request with invalid refresh token cookie
		c, rec := newTestContext(http.MethodPost, "/auth/refresh", nil)
		addCookies(c, &http.Cookie{Name: "refresh_token", Value: "invalid_refresh_token"})

		// Mock service response
		mockSessionService.refreshAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			return nil, errors.New("invalid refresh token")
		}

		// Execute
		err := controller.RefreshToken(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		checkResponseStatus(t, rec, http.StatusUnauthorized)

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid refresh token" {
			t.Errorf("Expected error 'Invalid refresh token', got '%s'", response["error"])
		}
	})
}

func TestGetUserProfile(t *testing.T) {
	// Table-driven tests for different authentication scenarios
	authTests := []struct {
		name           string
		cookies        []*http.Cookie
		mockTokenFunc  func(*MockSessionService, uuid.UUID)
		mockUserFunc   func(*MockUserService, uuid.UUID)
		expectedStatus int
		expectedError  string
		expectedCode   string
	}{
		{
			name: "Success",
			cookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			mockTokenFunc: func(mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return createTestSession(userID, token, "valid_refresh_token"), nil
				}
			},
			mockUserFunc: func(mockUser *MockUserService, userID uuid.UUID) {
				mockUser.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return &models.User{
						ID:            userID,
						Name:          "Test User",
						Email:         "test@example.com",
						EmailVerified: true,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Access Token",
			cookies:        []*http.Cookie{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Not authenticated",
		},
		{
			name: "Only Refresh Token",
			cookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Access token expired",
			expectedCode:   "token_expired",
		},
		{
			name: "Invalid Access Token",
			cookies: []*http.Cookie{
				{Name: "access_token", Value: "invalid_access_token"},
			},
			mockTokenFunc: func(mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid access token",
			expectedCode:   "token_invalid",
		},
		{
			name: "GetUserByID Failure",
			cookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			mockTokenFunc: func(mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return createTestSession(userID, token, "valid_refresh_token"), nil
				}
			},
			mockUserFunc: func(mockUser *MockUserService, userID uuid.UUID) {
				mockUser.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get user",
		},
	}

	for _, tc := range authTests {
		t.Run(tc.name, func(t *testing.T) {
			controller, mockUserService, mockSessionService := setupController()

			// Setup request
			c, rec := newTestContext(http.MethodGet, "/auth/profile", nil)
			if len(tc.cookies) > 0 {
				addCookies(c, tc.cookies...)
			}

			// Setup mocks if needed
			userID := uuid.New()
			if tc.mockTokenFunc != nil {
				tc.mockTokenFunc(mockSessionService, userID)
			}
			if tc.mockUserFunc != nil {
				tc.mockUserFunc(mockUserService, userID)
			}

			// Execute
			err := controller.GetUserProfile(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// For success case, verify user data
			if tc.expectedStatus == http.StatusOK {
				var user models.User
				err = json.Unmarshal(rec.Body.Bytes(), &user)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if user.ID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, user.ID)
				}
			} else {
				// For error cases, verify error response
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if tc.expectedError != "" && response["error"] != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, response["error"])
				}

				if tc.expectedCode != "" && response["code"] != tc.expectedCode {
					t.Errorf("Expected code '%s', got '%s'", tc.expectedCode, response["code"])
				}
			}
		})
	}
}

func TestGetUserProfileWithTrips(t *testing.T) {
	// Reuse the same table test structure as GetUserProfile with small modifications
	authTests := []struct {
		name           string
		cookies        []*http.Cookie
		limit          int
		offset         int
		mockTokenFunc  func(*MockSessionService, uuid.UUID)
		mockUserFunc   func(*MockUserService, uuid.UUID, int, int)
		expectedStatus int
		expectedError  string
		expectedCode   string
	}{
		{
			name: "Success",
			cookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			limit:  10,
			offset: 0,
			mockTokenFunc: func(mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return createTestSession(userID, token, "valid_refresh_token"), nil
				}
			},
			mockUserFunc: func(mockUser *MockUserService, userID uuid.UUID, limit, offset int) {
				mockUser.getUserWithTripsFunc = func(ctx context.Context, id uuid.UUID, l, o int) (*models.User, error) {
					if l != limit || o != offset {
						return nil, errors.New("unexpected pagination params")
					}
					return &models.User{
						ID:            userID,
						Name:          "Test User",
						Email:         "test@example.com",
						EmailVerified: true,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
						// Add trips data here as needed
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Access Token",
			cookies:        []*http.Cookie{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Not authenticated",
		},
		{
			name: "GetUserWithTrips Failure",
			cookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			mockTokenFunc: func(mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return createTestSession(userID, token, "valid_refresh_token"), nil
				}
			},
			mockUserFunc: func(mockUser *MockUserService, userID uuid.UUID, limit, offset int) {
				mockUser.getUserWithTripsFunc = func(ctx context.Context, id uuid.UUID, l, o int) (*models.User, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get user profile with trips",
		},
	}

	for _, tc := range authTests {
		t.Run(tc.name, func(t *testing.T) {
			controller, mockUserService, mockSessionService := setupController()

			// Setup request with query params for pagination if needed
			path := "/auth/profile/trips"
			if tc.limit > 0 || tc.offset > 0 {
				path = path + fmt.Sprintf("?limit=%d&offset=%d", tc.limit, tc.offset)
			}

			c, rec := newTestContext(http.MethodGet, path, nil)

			// Set query params in context
			if tc.limit > 0 {
				c.QueryParams().Set("limit", strconv.Itoa(tc.limit))
			}
			if tc.offset > 0 {
				c.QueryParams().Set("offset", strconv.Itoa(tc.offset))
			}

			if len(tc.cookies) > 0 {
				addCookies(c, tc.cookies...)
			}

			// Setup mocks if needed
			userID := uuid.New()
			if tc.mockTokenFunc != nil {
				tc.mockTokenFunc(mockSessionService, userID)
			}
			if tc.mockUserFunc != nil {
				tc.mockUserFunc(mockUserService, userID, tc.limit, tc.offset)
			}

			// Execute
			err := controller.GetUserProfileWithTrips(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// For success case, verify user data
			if tc.expectedStatus == http.StatusOK {
				var user models.User
				err = json.Unmarshal(rec.Body.Bytes(), &user)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if user.ID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, user.ID)
				}
			} else {
				// For error cases, verify error response
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if tc.expectedError != "" && response["error"] != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, response["error"])
				}

				if tc.expectedCode != "" && response["code"] != tc.expectedCode {
					t.Errorf("Expected code '%s', got '%s'", tc.expectedCode, response["code"])
				}
			}
		})
	}
}

func TestGetCSRFToken(t *testing.T) {
	controller, _, _ := setupController()

	// Setup request
	c, rec := newTestContext(http.MethodGet, "/auth/csrf", nil)

	// Set CSRF token in context
	c.Set("csrf", "test_csrf_token")

	// Execute
	err := controller.GetCSRFToken(c)
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
}
