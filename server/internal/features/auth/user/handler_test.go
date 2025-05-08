package user_test

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
	"black-lotus/internal/features/auth/user"
)

// MockRepository implements user.Repository for testing
type MockRepository struct {
	getUserByIDFunc func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *MockRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented")
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
		t.Errorf("Expected status %d, got %d", expectedStatus, rec.Code)
	}
}

// Setup creates handler with mock repositories for testing
func setupHandler() (user.HandlerInterface, *MockRepository, *MockSessionService) {
	mockRepo := &MockRepository{}
	mockSessionService := &MockSessionService{}

	// Create service
	service := user.NewService(mockRepo)

	// Create handler
	handler := user.NewHandler(service)

	return handler, mockRepo, mockSessionService
}

func TestHandlerGetUserByID(t *testing.T) {
	testCases := []struct {
		name           string
		userID         string // UUID as string to simulate path param
		setupMocks     func(*MockRepository, *MockSessionService, string) *models.User
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "SuccessfulFetch",
			userID: uuid.New().String(),
			setupMocks: func(mockRepo *MockRepository, mockSession *MockSessionService, userIDStr string) *models.User {
				userID, _ := uuid.Parse(userIDStr)
				testUser := &models.User{
					ID:            userID,
					Name:          "Test User",
					Email:         "test@example.com",
					EmailVerified: true,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}

				mockRepo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					if id == userID {
						return testUser, nil
					}
					return nil, errors.New("user not found")
				}

				return testUser
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "InvalidUUID",
			userID: "not-a-uuid",
			setupMocks: func(mockRepo *MockRepository, mockSession *MockSessionService, userIDStr string) *models.User {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "UserNotFound",
			userID: uuid.New().String(),
			setupMocks: func(mockRepo *MockRepository, mockSession *MockSessionService, userIDStr string) *models.User {
				mockRepo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, errors.New("user not found")
				}
				return nil
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:   "NilUserReturned",
			userID: uuid.New().String(),
			setupMocks: func(mockRepo *MockRepository, mockSession *MockSessionService, userIDStr string) *models.User {
				mockRepo.getUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return nil, nil
				}
				return nil
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockRepo, mockSession := setupHandler()

			// Setup request
			c, rec := newTestContext(http.MethodGet, "/api/users/"+tc.userID)
			c.SetParamNames("id")
			c.SetParamValues(tc.userID)

			// Set up the mocks
			expectedUser := tc.setupMocks(mockRepo, mockSession, tc.userID)

			// Execute
			err := handler.GetUserByID(c)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check status code
			checkResponseStatus(t, rec, tc.expectedStatus)

			// Verify response
			if tc.expectedError {
				var errorResponse map[string]string
				json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				if errorResponse["error"] == "" {
					t.Errorf("Expected error message in response, got none")
				}
			} else {
				var user models.User
				err = json.Unmarshal(rec.Body.Bytes(), &user)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				expectedUserID, _ := uuid.Parse(tc.userID)
				if user.ID != expectedUserID {
					t.Errorf("Expected user ID %s, got %s", expectedUserID, user.ID)
				}

				// Only check name if expectedUser is not nil
				if expectedUser != nil && user.Name != expectedUser.Name {
					t.Errorf("Expected user name %s, got %s", expectedUser.Name, user.Name)
				}
			}
		})
	}
}
