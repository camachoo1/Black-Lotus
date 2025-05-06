package controllers_test

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

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return s != "" && strings.Contains(s, substr)
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
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

func TestRegisterUser(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	t.Run("Successful Registration", func(t *testing.T) {
		// Create test input
		input := models.CreateUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
				return &models.Session{
					ID:            uuid.New(),
					UserID:        userID,
					AccessToken:   "test_access_token",
					RefreshToken:  "test_refresh_token",
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
					CreatedAt:     time.Now(),
				}, nil
			}
			return nil, errors.New("unexpected user ID")
		}

		// Execute
		err := controller.RegisterUser(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		// Verify response
		var response models.User
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, response.ID)
		}

		if response.Name != input.Name {
			t.Errorf("Expected name %s, got %s", input.Name, response.Name)
		}

		if response.Email != input.Email {
			t.Errorf("Expected email %s, got %s", input.Email, response.Email)
		}

		// Check if cookies are set
		cookies := rec.Result().Cookies()
		var accessTokenFound, refreshTokenFound bool
		for _, cookie := range cookies {
			if cookie.Name == "access_token" {
				accessTokenFound = true
				if cookie.Value != "test_access_token" {
					t.Errorf("Expected access token %s, got %s", "test_access_token", cookie.Value)
				}
			}
			if cookie.Name == "refresh_token" {
				refreshTokenFound = true
				if cookie.Value != "test_refresh_token" {
					t.Errorf("Expected refresh token %s, got %s", "test_refresh_token", cookie.Value)
				}
			}
		}

		if !accessTokenFound {
			t.Error("Access token cookie was not set")
		}

		if !refreshTokenFound {
			t.Error("Refresh token cookie was not set")
		}
	})

	t.Run("User Already Exists", func(t *testing.T) {
		// Create test input
		input := models.CreateUserInput{
			Name:     "Existing User",
			Email:    "existing@example.com",
			Password: stringPtr("Password123!"),
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		if rec.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, rec.Code)
		}

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

	t.Run("Validation Error", func(t *testing.T) {
		// Create invalid input (missing password)
		input := models.CreateUserInput{
			Name:  "Test User",
			Email: "test@example.com",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := controller.RegisterUser(c)
		if err != nil {
			t.Errorf("Expected no error from handler, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

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
			t.Error("Expected validation details, got nil or wrong type")
		}
	})
}

func TestLoginUser(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	t.Run("Successful Login", func(t *testing.T) {
		// Create test input
		input := models.LoginUserInput{
			Email:    "test@example.com",
			Password: "Password123!",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(inputJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
				return &models.Session{
					ID:            uuid.New(),
					UserID:        userID,
					AccessToken:   "test_access_token",
					RefreshToken:  "test_refresh_token",
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
					CreatedAt:     time.Now(),
				}, nil
			}
			return nil, errors.New("unexpected user ID")
		}

		// Execute
		err := controller.LoginUser(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Verify response
		var response models.User
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, response.ID)
		}

		if response.Name != loggedInUser.Name {
			t.Errorf("Expected name %s, got %s", loggedInUser.Name, response.Name)
		}

		if response.Email != input.Email {
			t.Errorf("Expected email %s, got %s", input.Email, response.Email)
		}

		// Check if cookies are set
		cookies := rec.Result().Cookies()
		var accessTokenFound, refreshTokenFound bool
		for _, cookie := range cookies {
			if cookie.Name == "access_token" {
				accessTokenFound = true
				if cookie.Value != "test_access_token" {
					t.Errorf("Expected access token %s, got %s", "test_access_token", cookie.Value)
				}
			}
			if cookie.Name == "refresh_token" {
				refreshTokenFound = true
				if cookie.Value != "test_refresh_token" {
					t.Errorf("Expected refresh token %s, got %s", "test_refresh_token", cookie.Value)
				}
			}
		}

		if !accessTokenFound {
			t.Error("Access token cookie was not set")
		}

		if !refreshTokenFound {
			t.Error("Refresh token cookie was not set")
		}
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		// Create test input
		input := models.LoginUserInput{
			Email:    "wrong@example.com",
			Password: "WrongPassword",
		}
		inputJSON, _ := json.Marshal(input)

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(inputJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

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
}

func TestLogoutUser(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	t.Run("Successful Logout", func(t *testing.T) {
		// Setup request with cookies
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		accessCookie := &http.Cookie{
			Name:  "access_token",
			Value: "test_access_token",
		}
		refreshCookie := &http.Cookie{
			Name:  "refresh_token",
			Value: "test_refresh_token",
		}
		req.AddCookie(accessCookie)
		req.AddCookie(refreshCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

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
		cookies := rec.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
				if cookie.Value != "" {
					t.Errorf("Expected cookie %s value to be empty, got '%s'", cookie.Name, cookie.Value)
				}
				if cookie.MaxAge >= 0 {
					t.Errorf("Expected cookie %s MaxAge to be negative, got %d", cookie.Name, cookie.MaxAge)
				}
			}
		}
	})

	t.Run("Already Logged Out", func(t *testing.T) {
		// Setup request without cookies
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := controller.LogoutUser(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

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
}

func TestRefreshToken(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	t.Run("Successful Token Refresh", func(t *testing.T) {
		// Setup request with refresh token cookie
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		refreshCookie := &http.Cookie{
			Name:  "refresh_token",
			Value: "valid_refresh_token",
		}
		req.AddCookie(refreshCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

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
		cookies := rec.Result().Cookies()
		var accessTokenFound bool
		for _, cookie := range cookies {
			if cookie.Name == "access_token" {
				accessTokenFound = true
				if cookie.Value != "new_access_token" {
					t.Errorf("Expected new access token 'new_access_token', got '%s'", cookie.Value)
				}
			}
		}

		if !accessTokenFound {
			t.Error("Access token cookie was not set")
		}
	})

	t.Run("No Refresh Token", func(t *testing.T) {
		// Setup request without refresh token cookie
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := controller.RefreshToken(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

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

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		// Setup request with invalid refresh token cookie
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		refreshCookie := &http.Cookie{
			Name:  "refresh_token",
			Value: "invalid_refresh_token",
		}
		req.AddCookie(refreshCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

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

// Additional test for RegisterUser with invalid email format
func TestRegisterUserWithInvalidEmail(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Create test input with invalid email
	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "not-an-email",
		Password: stringPtr("Password123!"),
	}
	inputJSON, _ := json.Marshal(input)

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := controller.RegisterUser(c)
	if err != nil {
		t.Errorf("Expected no error from handler, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Verify response contains validation error for email
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
		t.Error("Expected validation details, got nil or wrong type")
	}
}

// Test for RegisterUser with weak password
func TestRegisterUserWithWeakPassword(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Create test input with weak password
	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: stringPtr("weak"),
	}
	inputJSON, _ := json.Marshal(input)

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := controller.RegisterUser(c)
	if err != nil {
		t.Errorf("Expected no error from handler, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Verify response contains validation errors for password
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	errorMsg, ok := response["error"].(string)
	if !ok || errorMsg != "Validation failed" {
		t.Errorf("Expected error 'Validation failed', got '%v'", response["error"])
	}
}

// Test for RegisterUser with session creation failure
func TestRegisterUserWithSessionFailure(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Create test input
	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: stringPtr("Password123!"),
	}
	inputJSON, _ := json.Marshal(input)

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

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

	// Check status code - should still be created
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	// Verify user is returned despite session failure
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
}

// Test LoginUser with session creation failure
func TestLoginUserWithSessionFailure(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Create test input
	input := models.LoginUserInput{
		Email:    "test@example.com",
		Password: "Password123!",
	}
	inputJSON, _ := json.Marshal(input)

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

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

	// Check status code - should be internal server error
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Verify error response
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !containsSubstring(response["error"], "Failed to create session") {
		t.Errorf("Expected error message containing 'Failed to create session', got '%s'", response["error"])
	}
}

// Test LogoutUser with partial token failure
func TestLogoutUserWithPartialFailure(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Setup request with cookies
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	accessCookie := &http.Cookie{
		Name:  "access_token",
		Value: "test_access_token",
	}
	refreshCookie := &http.Cookie{
		Name:  "refresh_token",
		Value: "test_refresh_token",
	}
	req.AddCookie(accessCookie)
	req.AddCookie(refreshCookie)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock endSessionByAccessToken to fail
	mockSessionService.endSessionByAccessTokenFunc = func(ctx context.Context, token string) error {
		return errors.New("access token deletion failed")
	}

	// Mock endSessionByRefreshToken to succeed
	mockSessionService.endSessionByRefreshTokenFunc = func(ctx context.Context, token string) error {
		return nil
	}

	// Execute
	err := controller.LogoutUser(c)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code - should still be successful
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify response
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Successfully logged out" {
		t.Errorf("Expected message 'Successfully logged out', got '%s'", response["message"])
	}

	// Check if cookies are cleared despite token deletion failure
	cookies := rec.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
			if cookie.Value != "" {
				t.Errorf("Expected cookie %s value to be empty, got '%s'", cookie.Name, cookie.Value)
			}
			if cookie.MaxAge >= 0 {
				t.Errorf("Expected cookie %s MaxAge to be negative, got %d", cookie.Name, cookie.MaxAge)
			}
		}
	}
}

// Test for GetUserProfile
func TestGetUserProfile(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	t.Run("Success", func(t *testing.T) {
		// Setup request with access token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
		accessCookie := &http.Cookie{
			Name:  "access_token",
			Value: "valid_access_token",
		}
		req.AddCookie(accessCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock session service
		userID := uuid.New()
		mockSessionService.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			if token == "valid_access_token" {
				return &models.Session{
					ID:            uuid.New(),
					UserID:        userID,
					AccessToken:   token,
					RefreshToken:  "valid_refresh_token",
					AccessExpiry:  time.Now().Add(15 * time.Minute),
					RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
				}, nil
			}
			return nil, errors.New("invalid token")
		}

		// Mock user service
		mockUserService.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			if id == userID {
				return &models.User{
					ID:            userID,
					Name:          "Test User",
					Email:         "test@example.com",
					EmailVerified: true,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil
			}
			return nil, errors.New("user not found")
		}

		// Execute
		err := controller.GetUserProfile(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Verify response
		var user models.User
		err = json.Unmarshal(rec.Body.Bytes(), &user)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if user.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, user.ID)
		}
	})

	t.Run("No Access Token", func(t *testing.T) {
		// Setup request without access token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := controller.GetUserProfile(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Not authenticated" {
			t.Errorf("Expected error 'Not authenticated', got '%s'", response["error"])
		}
	})

	t.Run("Only Refresh Token", func(t *testing.T) {
		// Setup request with only refresh token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
		refreshCookie := &http.Cookie{
			Name:  "refresh_token",
			Value: "valid_refresh_token",
		}
		req.AddCookie(refreshCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := controller.GetUserProfile(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Access token expired" || response["code"] != "token_expired" {
			t.Errorf("Expected error 'Access token expired', got '%s'", response["error"])
		}
	})

	t.Run("Invalid Access Token", func(t *testing.T) {
		// Setup request with invalid access token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
		accessCookie := &http.Cookie{
			Name:  "access_token",
			Value: "invalid_access_token",
		}
		req.AddCookie(accessCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock session service
		mockSessionService.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			return nil, errors.New("invalid token")
		}

		// Execute
		err := controller.GetUserProfile(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid access token" || response["code"] != "token_invalid" {
			t.Errorf("Expected error 'Invalid access token', got '%s'", response["error"])
		}
	})

	t.Run("GetUserByID Failure", func(t *testing.T) {
		// Setup request with valid access token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
		accessCookie := &http.Cookie{
			Name:  "access_token",
			Value: "valid_access_token",
		}
		req.AddCookie(accessCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock session service success
		userID := uuid.New()
		mockSessionService.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			return &models.Session{
				ID:           uuid.New(),
				UserID:       userID,
				AccessToken:  token,
				RefreshToken: "valid_refresh_token",
			}, nil
		}

		// Mock user service failure
		mockUserService.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return nil, errors.New("database error")
		}

		// Execute
		err := controller.GetUserProfile(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Failed to get user" {
			t.Errorf("Expected error 'Failed to get user', got '%s'", response["error"])
		}
	})
}

// Test GetCSRFToken
func TestGetCSRFToken(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Setup request
	req := httptest.NewRequest(http.MethodGet, "/auth/csrf", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set CSRF token in context
	c.Set("csrf", "test_csrf_token")

	// Execute
	err := controller.GetCSRFToken(c)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

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

// Test GetUserProfileWithTrips
func TestGetUserProfileWithTrips(t *testing.T) {
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	t.Run("Success", func(t *testing.T) {
		// Setup request with access token and pagination params
		req := httptest.NewRequest(http.MethodGet, "/auth/profile/trips?limit=10&offset=0", nil)
		accessCookie := &http.Cookie{
			Name:  "access_token",
			Value: "valid_access_token",
		}
		req.AddCookie(accessCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock session service
		userID := uuid.New()
		mockSessionService.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			if token == "valid_access_token" {
				return &models.Session{
					ID:           uuid.New(),
					UserID:       userID,
					AccessToken:  token,
					RefreshToken: "valid_refresh_token",
				}, nil
			}
			return nil, errors.New("invalid token")
		}

		// Mock user service
		mockUserService.getUserWithTripsFunc = func(ctx context.Context, id uuid.UUID, limit, offset int) (*models.User, error) {
			if id == userID && limit == 10 && offset == 0 {
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
			return nil, errors.New("user not found or invalid pagination")
		}

		// Execute
		err := controller.GetUserProfileWithTrips(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Verify response
		var user models.User
		err = json.Unmarshal(rec.Body.Bytes(), &user)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if user.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, user.ID)
		}
	})

	t.Run("No Access Token", func(t *testing.T) {
		// Setup request without access token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile/trips", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := controller.GetUserProfileWithTrips(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Not authenticated" {
			t.Errorf("Expected error 'Not authenticated', got '%s'", response["error"])
		}
	})

	t.Run("GetUserWithTrips Failure", func(t *testing.T) {
		// Setup request with valid access token
		req := httptest.NewRequest(http.MethodGet, "/auth/profile/trips", nil)
		accessCookie := &http.Cookie{
			Name:  "access_token",
			Value: "valid_access_token",
		}
		req.AddCookie(accessCookie)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock session service success
		userID := uuid.New()
		mockSessionService.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
			return &models.Session{
				ID:           uuid.New(),
				UserID:       userID,
				AccessToken:  token,
				RefreshToken: "valid_refresh_token",
			}, nil
		}

		// Mock user service failure
		mockUserService.getUserWithTripsFunc = func(ctx context.Context, id uuid.UUID, limit, offset int) (*models.User, error) {
			return nil, errors.New("database error")
		}

		// Execute
		err := controller.GetUserProfileWithTrips(c)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check status code
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		// Verify response
		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Failed to get user profile with trips" {
			t.Errorf("Expected error 'Failed to get user profile with trips', got '%s'", response["error"])
		}
	})
}

// TestRegisterUserWithValidationErrors tests various validation scenarios
func TestRegisterUserWithValidationErrors(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	testCases := []struct {
		name          string
		input         models.CreateUserInput
		expectedField string // Exact field name as it appears in the response
		expectedMsg   string // Exact error message to match
	}{
		{
			name: "Missing Password",
			input: models.CreateUserInput{
				Name:  "Test User",
				Email: "test@example.com",
				// Password is nil
			},
			expectedField: "password",
			expectedMsg:   "password is required",
		},
		{
			name: "Invalid Email",
			input: models.CreateUserInput{
				Name:     "Test User",
				Email:    "not-an-email",
				Password: stringPtr("Password123!"),
			},
			expectedField: "email",
			expectedMsg:   "Please enter a valid email address",
		},
		{
			name: "Password Too Short",
			input: models.CreateUserInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: stringPtr("Pa1!"), // Too short
			},
			expectedField: "password",
			expectedMsg:   "password must be at least 6 characters long",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy of the test input to avoid changes between tests
			inputJSON, _ := json.Marshal(tc.input)

			// Setup request
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := controller.RegisterUser(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}

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

			// Get the error message for the exact field name
			fieldError, exists := details[tc.expectedField]
			if !exists {
				t.Errorf("Expected validation error for field '%s', but none found in: %v",
					tc.expectedField, details)
				return
			}

			// Check if the error message exactly matches the expected text
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
}

// TestRegisterUserWithPasswordValidationErrors tests each password validation rule separately
func TestRegisterUserWithPasswordValidationErrors(t *testing.T) {
	testCases := []struct {
		name        string
		password    string
		expectedMsg string // Exact error message to match
	}{
		{
			name:        "Password Without Uppercase",
			password:    "password123!",
			expectedMsg: "Password must contain at least one uppercase letter",
		},
		{
			name:        "Password Without Lowercase",
			password:    "PASSWORD123!",
			expectedMsg: "Password must contain at least one lowercase letter",
		},
		{
			name:        "Password Without Number",
			password:    "Password!",
			expectedMsg: "Password must contain at least one number",
		},
		{
			name:        "Password Without Special Character",
			password:    "Password123",
			expectedMsg: "Password must contain at least one special character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			mockUserService := &MockUserService{}
			mockSessionService := &MockSessionService{}
			controller := controllers.NewUserController(mockUserService, mockSessionService)

			// Create input with the test password
			input := models.CreateUserInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: stringPtr(tc.password),
			}
			inputJSON, _ := json.Marshal(input)

			// Setup request
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := controller.RegisterUser(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}

			// Verify response contains validation error
			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			details, ok := response["details"].(map[string]interface{})
			if !ok || details == nil {
				t.Errorf("Expected validation details, got nil or wrong type: %v", response)
				return
			}

			// Get the error message for the password field
			// Use lowercase "password" as the field name based on error messages
			fieldError, exists := details["password"]
			if !exists {
				t.Errorf("Expected validation error for field 'password', but none found in: %v", details)
				return
			}

			// Check if the error message exactly matches the expected text
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
}

// TestRegisterUserWithGeneralFailure tests the case where user creation fails with internal error
func TestRegisterUserWithGeneralFailure(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Create test input with valid data
	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: stringPtr("Password123!"),
	}
	inputJSON, _ := json.Marshal(input)

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

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
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Verify error response
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Failed to create user" {
		t.Errorf("Expected error 'Failed to create user', got '%s'", response["error"])
	}
}

// TestLogoutUserWithRefreshTokenFailure tests the case where ending the refresh token session fails
func TestLogoutUserWithRefreshTokenFailure(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Setup request with only refresh token
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	refreshCookie := &http.Cookie{
		Name:  "refresh_token",
		Value: "test_refresh_token",
	}
	req.AddCookie(refreshCookie)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

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
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

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
	cookies := rec.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			if cookie.Value != "" {
				t.Errorf("Expected cookie %s value to be empty, got '%s'", cookie.Name, cookie.Value)
			}
			if cookie.MaxAge >= 0 {
				t.Errorf("Expected cookie %s MaxAge to be negative, got %d", cookie.Name, cookie.MaxAge)
			}
		}
	}
}

// TestGetUserProfileWithTripsWithRefreshTokenOnly tests the case where access token is missing but refresh token exists
func TestGetUserProfileWithTripsWithRefreshTokenOnly(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Setup request with only refresh token
	req := httptest.NewRequest(http.MethodGet, "/auth/profile/trips", nil)
	refreshCookie := &http.Cookie{
		Name:  "refresh_token",
		Value: "valid_refresh_token",
	}
	req.AddCookie(refreshCookie)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := controller.GetUserProfileWithTrips(c)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	// Verify response
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Access token expired" || response["code"] != "token_expired" {
		t.Errorf("Expected error 'Access token expired', got '%s' with code '%s'",
			response["error"], response["code"])
	}
}

// TestGetUserProfileWithTripsWithInvalidToken tests the case where access token is invalid
func TestGetUserProfileWithTripsWithInvalidToken(t *testing.T) {
	// Setup
	e := echo.New()
	mockUserService := &MockUserService{}
	mockSessionService := &MockSessionService{}
	controller := controllers.NewUserController(mockUserService, mockSessionService)

	// Setup request with invalid access token
	req := httptest.NewRequest(http.MethodGet, "/auth/profile/trips", nil)
	accessCookie := &http.Cookie{
		Name:  "access_token",
		Value: "invalid_access_token",
	}
	req.AddCookie(accessCookie)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock session service to return an error for invalid token
	mockSessionService.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
		return nil, errors.New("invalid token")
	}

	// Execute
	err := controller.GetUserProfileWithTrips(c)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	// Verify response
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Invalid access token" || response["code"] != "token_invalid" {
		t.Errorf("Expected error 'Invalid access token', got '%s' with code '%s'",
			response["error"], response["code"])
	}
}
