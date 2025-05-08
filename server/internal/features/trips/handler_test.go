package trips_test

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
	"black-lotus/internal/features/trips"
)

// MockTripService implements trips.ServiceInterface for testing
type MockTripService struct {
	createTripFunc       func(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	updateTripFunc       func(ctx context.Context, tripID uuid.UUID, userID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	deleteTripFunc       func(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) error
	getTripByIDFunc      func(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (*models.Trip, error)
	getTripWithUserFunc  func(ctx context.Context, tripID uuid.UUID, requestUserID uuid.UUID) (*models.Trip, error)
	getUserWithTripsFunc func(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error)
	getTripsByUserIDFunc func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error)
}

func (m *MockTripService) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	if m.createTripFunc != nil {
		return m.createTripFunc(ctx, userID, input)
	}
	return nil, errors.New("CreateTrip not implemented")
}

func (m *MockTripService) UpdateTrip(ctx context.Context, tripID uuid.UUID, userID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	if m.updateTripFunc != nil {
		return m.updateTripFunc(ctx, tripID, userID, input)
	}
	return nil, errors.New("UpdateTrip not implemented")
}

func (m *MockTripService) DeleteTrip(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) error {
	if m.deleteTripFunc != nil {
		return m.deleteTripFunc(ctx, tripID, userID)
	}
	return errors.New("DeleteTrip not implemented")
}

func (m *MockTripService) GetTripByID(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (*models.Trip, error) {
	if m.getTripByIDFunc != nil {
		return m.getTripByIDFunc(ctx, tripID, userID)
	}
	return nil, errors.New("GetTripByID not implemented")
}

func (m *MockTripService) GetTripWithUser(ctx context.Context, tripID uuid.UUID, requestUserID uuid.UUID) (*models.Trip, error) {
	if m.getTripWithUserFunc != nil {
		return m.getTripWithUserFunc(ctx, tripID, requestUserID)
	}
	return nil, errors.New("GetTripWithUser not implemented")
}

func (m *MockTripService) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error) {
	if m.getUserWithTripsFunc != nil {
		return m.getUserWithTripsFunc(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetUserWithTrips not implemented")
}

func (m *MockTripService) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error) {
	if m.getTripsByUserIDFunc != nil {
		return m.getTripsByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetTripsByUserID not implemented")
}

// MockSessionService implements session.ServiceInterface for testing
type MockSessionService struct {
	validateAccessTokenFunc func(ctx context.Context, token string) (*models.Session, error)
}

func (m *MockSessionService) ValidateAccessToken(ctx context.Context, token string) (*models.Session, error) {
	if m.validateAccessTokenFunc != nil {
		return m.validateAccessTokenFunc(ctx, token)
	}
	return nil, errors.New("ValidateAccessToken not implemented")
}

func (m *MockSessionService) ValidateRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	return nil, errors.New("ValidateRefreshToken not implemented")
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	return nil, errors.New("CreateSession not implemented")
}

func (m *MockSessionService) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	return nil, errors.New("RefreshAccessToken not implemented")
}

func (m *MockSessionService) EndSessionByAccessToken(ctx context.Context, token string) error {
	return errors.New("EndSessionByAccessToken not implemented")
}

func (m *MockSessionService) EndSessionByRefreshToken(ctx context.Context, token string) error {
	return errors.New("EndSessionByRefreshToken not implemented")
}

func (m *MockSessionService) EndAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return errors.New("EndAllUserSessions not implemented")
}

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

// CreateTestSession creates a test session
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

// Helper function to setup service for testing
func setupHandlerTest() (*trips.Handler, *MockTripService, *MockSessionService) {
	mockService := &MockTripService{}
	mockSessionService := &MockSessionService{}

	// Set default implementations for the mock service
	mockService.createTripFunc = func(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
		return &models.Trip{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        input.Name,
			Description: input.Description,
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Location:    input.Location,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}

	mockService.getTripByIDFunc = func(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (*models.Trip, error) {
		return &models.Trip{
			ID:          tripID,
			UserID:      userID,
			Name:        "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(7 * 24 * time.Hour),
			Location:    "Test City",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}

	mockService.updateTripFunc = func(ctx context.Context, tripID uuid.UUID, userID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
		// Create a base trip
		trip := &models.Trip{
			ID:          tripID,
			UserID:      userID,
			Name:        "Original Trip",
			Description: "Original Description",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(7 * 24 * time.Hour),
			Location:    "Original City",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Update fields based on input
		if input.Name != nil {
			trip.Name = *input.Name
		}
		if input.Description != nil {
			trip.Description = *input.Description
		}
		if input.StartDate != nil {
			trip.StartDate = *input.StartDate
		}
		if input.EndDate != nil {
			trip.EndDate = *input.EndDate
		}
		if input.Location != nil {
			trip.Location = *input.Location
		}

		return trip, nil
	}

	mockService.deleteTripFunc = func(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) error {
		return nil
	}

	mockService.getTripsByUserIDFunc = func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error) {
		return []*models.Trip{
			{
				ID:          uuid.New(),
				UserID:      userID,
				Name:        "Trip 1",
				Description: "Description 1",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Location 1",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				UserID:      userID,
				Name:        "Trip 2",
				Description: "Description 2",
				StartDate:   time.Now().Add(14 * 24 * time.Hour),
				EndDate:     time.Now().Add(21 * 24 * time.Hour),
				Location:    "Location 2",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}, nil
	}

	handler := trips.NewHandler(mockService, mockSessionService)
	return handler, mockService, mockSessionService
}

func TestHandlerCreateTrip(t *testing.T) {
	testCases := []struct {
		name           string
		input          models.CreateTripInput
		setupCookies   []*http.Cookie
		setupMocks     func(*testing.T, *MockTripService, *MockSessionService, uuid.UUID)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "SuccessfulCreation",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.createTripFunc = func(ctx context.Context, uid uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
					return &models.Trip{
						ID:          uuid.New(),
						UserID:      uid,
						Name:        input.Name,
						Description: input.Description,
						StartDate:   input.StartDate,
						EndDate:     input.EndDate,
						Location:    input.Location,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "NoAccessToken",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupCookies: []*http.Cookie{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No mocks needed as request will fail early
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "InvalidAccessToken",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "invalid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "InvalidInputValidation",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				// Missing StartDate
				// Missing EndDate
				Location: "Test City",
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "ServiceError",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				// Override the default implementation specifically for this test case
				mockService.createTripFunc = func(ctx context.Context, uid uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:  "InvalidRequestBody",
			input: models.CreateTripInput{},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:  "AccessTokenExpired",
			input: models.CreateTripInput{},
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No mocks needed as we're testing the token expired path
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:  "InvalidRequestBodyFormat",
			input: models.CreateTripInput{}, // This won't matter because we'll use invalid JSON
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "ValidationFailure_WithCustomField",
			input: models.CreateTripInput{
				// Missing required fields
				Name:        "Test Trip",
				Description: "Test Description",
				// Missing StartDate, EndDate, Location
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "NonValidationError",
			input: models.CreateTripInput{
				Name:        "Test Trip",
				Description: "Test Description",
				StartDate:   time.Now().Add(24 * time.Hour),
				EndDate:     time.Now().Add(7 * 24 * time.Hour),
				Location:    "Test City",
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				// Instead of mocking the service, we'll use a custom validator that returns a non-ValidationErrors type
				mockService.createTripFunc = func(ctx context.Context, uid uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
					return nil, errors.New("internal error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:  "RefreshTokenOnly",
			input: models.CreateTripInput{},
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No mocks needed - testing token expired path
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockService, mockSession := setupHandlerTest()
			userID := uuid.New()

			if tc.name == "InvalidRequestBodyFormat" {
				// Setup request with invalid JSON
				invalidJSON := []byte(`{"name": "Test Trip" "description": "Invalid JSON format"}`) // Missing comma
				c, rec := newTestContext(http.MethodPost, "/api/trips", invalidJSON)

				// Add cookies
				addCookies(c, tc.setupCookies...)

				// Setup mocks
				tc.setupMocks(t, mockService, mockSession, userID)

				// Execute
				err := handler.CreateTrip(c)
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Check status code
				if rec.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
				}

				// Verify response body
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)

				if response["error"] != "Invalid request body" {
					t.Errorf("Expected error message 'Invalid request body', got '%s'", response["error"])
				}
			}

			// Create request body
			inputJSON, _ := json.Marshal(tc.input)

			// Setup request
			c, rec := newTestContext(http.MethodPost, "/api/trips", inputJSON)

			// Add cookies
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Setup mocks
			tc.setupMocks(t, mockService, mockSession, userID)

			// Execute
			err := handler.CreateTrip(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			if !tc.expectedError {
				var trip models.Trip
				err = json.Unmarshal(rec.Body.Bytes(), &trip)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if trip.Name != tc.input.Name {
					t.Errorf("Expected trip name '%s', got '%s'", tc.input.Name, trip.Name)
				}
			} else {
				var errorResponse map[string]interface{}
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)

				if errorResponse["error"] == nil && errorResponse["details"] == nil {
					t.Error("Expected error message in response")
				}
			}
		})
	}
}

func TestHandlerGetTrip(t *testing.T) {
	testCases := []struct {
		name           string
		setupCookies   []*http.Cookie
		setupMocks     func(*testing.T, *MockTripService, *MockSessionService, uuid.UUID, uuid.UUID)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "SuccessfulRetrieval",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getTripByIDFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID) (*models.Trip, error) {
					if tid == tripID && uid == userID {
						return &models.Trip{
							ID:          tripID,
							UserID:      userID,
							Name:        "Test Trip",
							Description: "Test Description",
							StartDate:   time.Now().Add(24 * time.Hour),
							EndDate:     time.Now().Add(7 * 24 * time.Hour),
							Location:    "Test City",
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						}, nil
					}
					return nil, errors.New("trip not found")
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:         "NoAccessToken",
			setupCookies: []*http.Cookie{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				// No mocks needed as request will fail early
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "InvalidAccessToken",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "invalid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "TripNotFound",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getTripByIDFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID) (*models.Trip, error) {
					return nil, errors.New("trip not found")
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name: "InvalidTripID",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},

		{
			name: "InternalServerError",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getTripByIDFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID) (*models.Trip, error) {
					// Return an error that's not "trip not found" to trigger the internal server error path
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},

		{
			name: "RefreshTokenOnly",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				// No mocks needed
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockService, mockSession := setupHandlerTest()
			tripID := uuid.New()
			userID := uuid.New()
			if tc.name == "InvalidTripID" {
				// Setup request with invalid UUID
				c, rec := newTestContext(http.MethodGet, "/api/trips/not-a-valid-uuid", nil)
				c.SetParamNames("id")
				c.SetParamValues("not-a-valid-uuid")

				// Add cookies
				if len(tc.setupCookies) > 0 {
					addCookies(c, tc.setupCookies...)
				}

				// Setup mocks
				tc.setupMocks(t, mockService, mockSession, tripID, userID)

				// Execute
				err := handler.GetTrip(c)
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Check status code
				checkResponseStatus(t, rec, tc.expectedStatus)

				// Verify error response
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)

				if errorResponse["error"] != "Invalid trip ID" {
					t.Errorf("Expected error message 'Invalid trip ID', got '%s'", errorResponse["error"])
				}

				// Skip the rest of the test
				return
			}

			// Setup request
			c, rec := newTestContext(http.MethodGet, "/api/trips/"+tripID.String(), nil)
			c.SetParamNames("id")
			c.SetParamValues(tripID.String())

			// Add cookies
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Setup mocks
			tc.setupMocks(t, mockService, mockSession, tripID, userID)

			// Execute
			err := handler.GetTrip(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			if !tc.expectedError {
				var trip models.Trip
				err = json.Unmarshal(rec.Body.Bytes(), &trip)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if trip.ID != tripID {
					t.Errorf("Expected trip ID %s, got %s", tripID, trip.ID)
				}
			} else {
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)

				if errorResponse["error"] == "" {
					t.Error("Expected error message in response")
				}
			}
		})
	}
}

func TestHandlerUpdateTrip(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name           string
		updateInput    models.UpdateTripInput
		setupCookies   []*http.Cookie
		setupMocks     func(*testing.T, *MockTripService, *MockSessionService, uuid.UUID, uuid.UUID)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "SuccessfulUpdate",
			updateInput: models.UpdateTripInput{
				Name:        stringPtr("Updated Trip"),
				Description: stringPtr("Updated Description"),
				StartDate:   timePtr(now.Add(24 * time.Hour)),
				EndDate:     timePtr(now.Add(96 * time.Hour)),
				Location:    stringPtr("Updated City"),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					if tid == tripID && uid == userID {
						return &models.Trip{
							ID:          tripID,
							UserID:      userID,
							Name:        *input.Name,
							Description: *input.Description,
							StartDate:   *input.StartDate,
							EndDate:     *input.EndDate,
							Location:    *input.Location,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						}, nil
					}
					return nil, errors.New("trip not found")
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "NoAccessToken",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupCookies: []*http.Cookie{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				// No mocks needed as request will fail early
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "TripNotFound",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					return nil, errors.New("trip not found")
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name: "UnauthorizedAccess",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					return nil, errors.New("unauthorized access to trip")
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  true,
		},
		{
			name:        "InvalidRequestBody",
			updateInput: models.UpdateTripInput{}, // Empty input
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				// Important: Override the default implementation to ensure it's not called
				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					t.Error("updateTripFunc should not be called for empty input")
					return nil, errors.New("should not be called")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "ValidationError",
			updateInput: models.UpdateTripInput{
				// Invalid input that would fail validation
				StartDate: timePtr(time.Now().Add(48 * time.Hour)),
				EndDate:   timePtr(time.Now().Add(24 * time.Hour)),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				// Return validation error
				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					return nil, errors.New("end date cannot be before start date")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "AccessTokenExpired",
			updateInput: models.UpdateTripInput{},
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				// No mocks needed as we're testing the token expired path
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "InvalidTripID",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "InvalidAccessToken",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "invalid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},

		{
			name:        "InvalidRequestBodyFormat",
			updateInput: models.UpdateTripInput{}, // This won't matter because we'll use invalid JSON
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},

		{
			name: "ValidationError",
			updateInput: models.UpdateTripInput{
				StartDate: timePtr(time.Now().Add(48 * time.Hour)),
				EndDate:   timePtr(time.Now().Add(24 * time.Hour)),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					return nil, errors.New("end date cannot be before start date")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},

		{
			name: "NonValidationError",
			updateInput: models.UpdateTripInput{
				Name: stringPtr("Updated Trip"),
			},
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.updateTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
					return nil, errors.New("some other error")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockService, mockSession := setupHandlerTest()
			tripID := uuid.New()
			userID := uuid.New()

			if tc.name == "InvalidRequestBodyFormat" {
				// Setup request with invalid JSON
				invalidJSON := []byte(`{"name": "Updated Trip" "description": "Invalid JSON format"}`) // Missing comma
				c, rec := newTestContext(http.MethodPut, "/api/trips/"+tripID.String(), invalidJSON)
				c.SetParamNames("id")
				c.SetParamValues(tripID.String())

				// Add cookies
				addCookies(c, tc.setupCookies...)

				// Setup mocks
				tc.setupMocks(t, mockService, mockSession, tripID, userID)

				// Execute
				err := handler.UpdateTrip(c)
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Check status code
				checkResponseStatus(t, rec, http.StatusBadRequest)

				// Verify response body
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)

				if response["error"] != "Invalid request body" {
					t.Errorf("Expected error message 'Invalid request body', got '%s'", response["error"])
				}
				return // Skip the rest of the test
			}

			if tc.name == "InvalidTripID" {
				// Setup request with invalid UUID
				c, rec := newTestContext(http.MethodPut, "/api/trips/not-a-uuid", nil)
				c.SetParamNames("id")
				c.SetParamValues("not-a-uuid")

				// Add cookies
				if len(tc.setupCookies) > 0 {
					addCookies(c, tc.setupCookies...)
				}

				// Setup mocks
				tc.setupMocks(t, mockService, mockSession, tripID, userID)

				// Execute
				err := handler.UpdateTrip(c)
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Check status code
				checkResponseStatus(t, rec, tc.expectedStatus)

				// Verify error response
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)

				if response["error"] != "Invalid trip ID" {
					t.Errorf("Expected error message 'Invalid trip ID', got '%s'", response["error"])
				}
				return // Skip the rest of the test
			}

			// Create request body for normal test cases
			inputJSON, _ := json.Marshal(tc.updateInput)
			c, rec := newTestContext(http.MethodPut, "/api/trips/"+tripID.String(), inputJSON)
			c.SetParamNames("id")
			c.SetParamValues(tripID.String())

			// Add cookies
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Setup mocks
			tc.setupMocks(t, mockService, mockSession, tripID, userID)

			// Execute
			err := handler.UpdateTrip(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			if !tc.expectedError {
				var trip models.Trip
				err = json.Unmarshal(rec.Body.Bytes(), &trip)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if trip.ID != tripID {
					t.Errorf("Expected trip ID %s, got %s", tripID, trip.ID)
				}

				if tc.updateInput.Name != nil && trip.Name != *tc.updateInput.Name {
					t.Errorf("Expected trip name '%s', got '%s'", *tc.updateInput.Name, trip.Name)
				}
			} else {
				// For error cases, verify the error message
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}

				// Check that there is an error message
				errorMsg, ok := response["error"]
				if !ok || errorMsg == nil {
					t.Errorf("Expected error message in response, got %v", response)
				}
			}
		})
	}
}

func TestHandlerDeleteTrip(t *testing.T) {
	testCases := []struct {
		name           string
		setupCookies   []*http.Cookie
		setupMocks     func(*testing.T, *MockTripService, *MockSessionService, uuid.UUID, uuid.UUID)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "SuccessfulDelete",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.deleteTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID) error {
					if tid == tripID && uid == userID {
						return nil
					}
					return errors.New("trip not found")
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:         "NoAccessToken",
			setupCookies: []*http.Cookie{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				// No mocks needed as request will fail early
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "TripNotFound",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.deleteTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID) error {
					return errors.New("trip not found")
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name: "UnauthorizedAccess",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.deleteTripFunc = func(ctx context.Context, tid uuid.UUID, uid uuid.UUID) error {
					return errors.New("unauthorized access to trip")
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  true,
		},
		{
			name: "InvalidTripID",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "AccessTokenExpired",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				// No mocks needed as we're testing the token expired path
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "InvalidAccessToken",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "invalid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, tripID, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockService, mockSession := setupHandlerTest()
			tripID := uuid.New()
			userID := uuid.New()

			if tc.name == "InvalidTripID" {
				// Setup request with invalid UUID
				c, rec := newTestContext(http.MethodDelete, "/api/trips/not-a-uuid", nil)
				c.SetParamNames("id")
				c.SetParamValues("not-a-uuid")

				// Add cookies
				if len(tc.setupCookies) > 0 {
					addCookies(c, tc.setupCookies...)
				}

				// Setup mocks
				tc.setupMocks(t, mockService, mockSession, tripID, userID)

				// Execute
				err := handler.DeleteTrip(c)
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Check status code
				checkResponseStatus(t, rec, tc.expectedStatus)

				// Verify error response
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)

				if errorResponse["error"] != "Invalid trip ID" {
					t.Errorf("Expected error message 'Invalid trip ID', got '%s'", errorResponse["error"])
				}
				return // Skip the rest of the test
			}

			// Setup request
			c, rec := newTestContext(http.MethodDelete, "/api/trips/"+tripID.String(), nil)
			c.SetParamNames("id")
			c.SetParamValues(tripID.String())

			// Add cookies
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Setup mocks
			tc.setupMocks(t, mockService, mockSession, tripID, userID)

			// Execute
			err := handler.DeleteTrip(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Verify response
			if !tc.expectedError {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["message"] != "Trip deleted successfully" {
					t.Errorf("Expected success message, got: %s", response["message"])
				}
			} else {
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)

				if errorResponse["error"] == "" {
					t.Error("Expected error message in response")
				}
			}
		})
	}
}

func TestHandlerGetUserTrips(t *testing.T) {
	testCases := []struct {
		name           string
		setupCookies   []*http.Cookie
		queryParams    map[string]string
		setupMocks     func(*testing.T, *MockTripService, *MockSessionService, uuid.UUID)
		expectedStatus int
		expectedError  bool
		tripCount      int
	}{
		{
			name: "SuccessfulRetrieval",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			queryParams: map[string]string{
				"limit":  "10",
				"offset": "0",
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getTripsByUserIDFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					if uid == userID && limit == 10 && offset == 0 {
						return []*models.Trip{
							{
								ID:          uuid.New(),
								UserID:      userID,
								Name:        "Trip 1",
								Description: "Description 1",
								StartDate:   time.Now().Add(24 * time.Hour),
								EndDate:     time.Now().Add(7 * 24 * time.Hour),
								Location:    "Location 1",
							},
							{
								ID:          uuid.New(),
								UserID:      userID,
								Name:        "Trip 2",
								Description: "Description 2",
								StartDate:   time.Now().Add(14 * 24 * time.Hour),
								EndDate:     time.Now().Add(21 * 24 * time.Hour),
								Location:    "Location 2",
							},
						}, nil
					}
					return nil, errors.New("invalid parameters")
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			tripCount:      2,
		},
		{
			name:         "NoAccessToken",
			setupCookies: []*http.Cookie{},
			queryParams:  map[string]string{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No mocks needed as request will fail early
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			tripCount:      0,
		},
		{
			name: "InvalidAccessToken",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "invalid_access_token"},
			},
			queryParams: map[string]string{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			tripCount:      0,
		},
		{
			name: "ServiceError",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			queryParams: map[string]string{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getTripsByUserIDFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
			tripCount:      0,
		},
		{
			name: "EmptyTripList",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			queryParams: map[string]string{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getTripsByUserIDFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*models.Trip, error) {
					return []*models.Trip{}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			tripCount:      0,
		},
		{
			name: "RefreshTokenOnly",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			queryParams: map[string]string{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No mocks needed
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			tripCount:      0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockService, mockSession := setupHandlerTest()
			userID := uuid.New()

			// Setup request with query parameters
			c, rec := newTestContext(http.MethodGet, "/api/trips", nil)
			q := c.Request().URL.Query()
			for key, value := range tc.queryParams {
				q.Add(key, value)
			}
			c.Request().URL.RawQuery = q.Encode()

			// Add cookies
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Setup mocks
			tc.setupMocks(t, mockService, mockSession, userID)

			// Execute
			err := handler.GetUserTrips(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			if !tc.expectedError {
				var trips []*models.Trip
				err = json.Unmarshal(rec.Body.Bytes(), &trips)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(trips) != tc.tripCount {
					t.Errorf("Expected %d trips, got %d", tc.tripCount, len(trips))
				}
			} else {
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)

				if errorResponse["error"] == "" {
					t.Error("Expected error message in response")
				}
			}
		})
	}
}
