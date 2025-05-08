// internal/features/profiles/trips/handler_test.go
package trips_test

import (
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
	"black-lotus/internal/features/profiles/trips"
)

type MockTripService struct {
	getUserWithTripsFunc func(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error)
}

// Implement necessary methods to match the real service
func (m *MockTripService) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error) {
	if m.getUserWithTripsFunc != nil {
		return m.getUserWithTripsFunc(ctx, userID, limit, offset)
	}
	return nil, errors.New("GetUserWithTrips not implemented")
}

// MockSessionService implements session.ServiceInterface
type MockSessionService struct {
	validateAccessTokenFunc func(ctx context.Context, token string) (*models.Session, error)
}

// Implement session.ServiceInterface methods
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
func newTestContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
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

// Setup creates handler with mock service for testing
func setupHandler() (*trips.Handler, *MockTripService, *MockSessionService) {
	mockService := &MockTripService{}
	mockSessionService := &MockSessionService{}

	// Create handler
	handler := trips.NewHandler(mockService, mockSessionService)

	return handler, mockService, mockSessionService
}

func TestGetUserProfileWithTrips(t *testing.T) {
	testCases := []struct {
		name           string
		setupCookies   []*http.Cookie
		setupMocks     func(*testing.T, *MockTripService, *MockSessionService, uuid.UUID)
		expectedStatus int
		expectedError  bool
		tripCount      int
	}{
		{
			name: "SuccessfulFetch",
			setupCookies: []*http.Cookie{
				{Name: "access_token", Value: "valid_access_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// Mock session service to validate access token
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				// Mock trip service to return user with trips
				mockService.getUserWithTripsFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) (*models.User, error) {
					if uid == userID {
						return &models.User{
							ID:            userID,
							Name:          "Test User",
							Email:         "test@example.com",
							EmailVerified: true,
							Trips: []*models.Trip{
								{
									ID:          uuid.New(),
									UserID:      userID,
									Name:        "Trip to Paris",
									Description: "Vacation in Paris",
									Location:    "Paris",
								},
								{
									ID:          uuid.New(),
									UserID:      userID,
									Name:        "Trip to Rome",
									Description: "Business trip to Rome",
									Location:    "Rome",
								},
							},
						}, nil
					}
					return nil, errors.New("user not found")
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			tripCount:      2,
		},
		{
			name:         "NoAccessToken",
			setupCookies: []*http.Cookie{},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No need to setup mocks as this should fail early
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
			name: "TokenExpired",
			setupCookies: []*http.Cookie{
				{Name: "refresh_token", Value: "valid_refresh_token"},
			},
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				// No need to setup mocks as this should fail early due to missing access token
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
			setupMocks: func(t *testing.T, mockService *MockTripService, mockSession *MockSessionService, userID uuid.UUID) {
				mockSession.validateAccessTokenFunc = func(ctx context.Context, token string) (*models.Session, error) {
					if token == "valid_access_token" {
						return createTestSession(userID, token, "valid_refresh_token"), nil
					}
					return nil, errors.New("invalid token")
				}

				mockService.getUserWithTripsFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) (*models.User, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
			tripCount:      0,
		},
		{
			name: "UserNotFound",
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

				mockService.getUserWithTripsFunc = func(ctx context.Context, uid uuid.UUID, limit, offset int) (*models.User, error) {
					return nil, nil
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
			tripCount:      0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockService, mockSession := setupHandler()
			userID := uuid.New()

			// Setup request with query parameters
			c, rec := newTestContext(http.MethodGet, "/api/profile/trips?limit=10&offset=0")

			// Add cookies
			if len(tc.setupCookies) > 0 {
				addCookies(c, tc.setupCookies...)
			}

			// Setup mocks
			tc.setupMocks(t, mockService, mockSession, userID)

			// Execute
			err := handler.GetUserProfileWithTrips(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			if tc.expectedError {
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				if tc.expectedStatus != http.StatusOK && errorResponse["error"] == "" {
					t.Errorf("Expected error message in response, got none")
				}
			} else {
				var user models.User
				err = json.Unmarshal(rec.Body.Bytes(), &user)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if user.ID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, user.ID)
				}

				if user.Trips == nil {
					t.Error("Expected trips array, got nil")
				}

				if len(user.Trips) != tc.tripCount {
					t.Errorf("Expected %d trips, got %d", tc.tripCount, len(user.Trips))
				}

				if tc.tripCount > 0 {
					// Check trip names if we expect trips
					tripNames := []string{"Trip to Paris", "Trip to Rome"}
					for i, trip := range user.Trips {
						if trip.Name != tripNames[i] {
							t.Errorf("Expected trip %d to be named %s, got %s", i, tripNames[i], trip.Name)
						}
					}
				}
			}
		})
	}
}
