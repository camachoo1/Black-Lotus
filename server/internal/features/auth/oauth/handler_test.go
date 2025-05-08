// internal/features/auth/oauth/handler_test.go
package oauth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"black-lotus/internal/features/auth/oauth"
	"black-lotus/internal/features/auth/oauth/github"
	"black-lotus/internal/features/auth/oauth/google"
)

// MockGitHubHandler mocks the GitHub handler
type MockGitHubHandler struct {
	getAuthURLCalled     bool
	handleCallbackCalled bool
}

func (m *MockGitHubHandler) GetAuthURL(ctx echo.Context) error {
	m.getAuthURLCalled = true
	return ctx.String(http.StatusOK, "GitHub Auth URL")
}

func (m *MockGitHubHandler) HandleCallback(ctx echo.Context) error {
	m.handleCallbackCalled = true
	return ctx.String(http.StatusOK, "GitHub Callback")
}

// MockGoogleHandler mocks the Google handler
type MockGoogleHandler struct {
	getAuthURLCalled     bool
	handleCallbackCalled bool
}

func (m *MockGoogleHandler) GetAuthURL(ctx echo.Context) error {
	m.getAuthURLCalled = true
	return ctx.String(http.StatusOK, "Google Auth URL")
}

func (m *MockGoogleHandler) HandleCallback(ctx echo.Context) error {
	m.handleCallbackCalled = true
	return ctx.String(http.StatusOK, "Google Callback")
}

// Compile-time interface checks
var _ github.HandlerInterface = (*MockGitHubHandler)(nil)
var _ google.HandlerInterface = (*MockGoogleHandler)(nil)

// Helper function to create a new test context
func newTestContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// TestHandlerDelegation tests that the main OAuth handler correctly delegates to provider handlers
func TestHandlerDelegation(t *testing.T) {
	// Test cases for delegation
	tests := []struct {
		name           string
		method         string
		path           string
		handlerMethod  func(*oauth.Handler, echo.Context) error
		checkMock      func(*MockGitHubHandler, *MockGoogleHandler) bool
		expectedStatus int
	}{
		{
			name:   "GitHub Auth URL Delegation",
			method: http.MethodGet,
			path:   "/api/auth/github",
			handlerMethod: func(h *oauth.Handler, c echo.Context) error {
				return h.GetGitHubAuthURL(c)
			},
			checkMock: func(gh *MockGitHubHandler, g *MockGoogleHandler) bool {
				return gh.getAuthURLCalled
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "GitHub Callback Delegation",
			method: http.MethodGet,
			path:   "/api/auth/github/callback",
			handlerMethod: func(h *oauth.Handler, c echo.Context) error {
				return h.HandleGitHubCallback(c)
			},
			checkMock: func(gh *MockGitHubHandler, g *MockGoogleHandler) bool {
				return gh.handleCallbackCalled
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Google Auth URL Delegation",
			method: http.MethodGet,
			path:   "/api/auth/google",
			handlerMethod: func(h *oauth.Handler, c echo.Context) error {
				return h.GetGoogleAuthURL(c)
			},
			checkMock: func(gh *MockGitHubHandler, g *MockGoogleHandler) bool {
				return g.getAuthURLCalled
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Google Callback Delegation",
			method: http.MethodGet,
			path:   "/api/auth/google/callback",
			handlerMethod: func(h *oauth.Handler, c echo.Context) error {
				return h.HandleGoogleCallback(c)
			},
			checkMock: func(gh *MockGitHubHandler, g *MockGoogleHandler) bool {
				return g.handleCallbackCalled
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockGitHubHandler := &MockGitHubHandler{}
			mockGoogleHandler := &MockGoogleHandler{}

			// Create handler under test
			handler := oauth.NewHandler(mockGitHubHandler, mockGoogleHandler)

			// Setup Echo context
			c, rec := newTestContext(tc.method, tc.path)

			// Execute the handler method
			err := tc.handlerMethod(handler, c)

			// Verify
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check if the correct method was called
			if !tc.checkMock(mockGitHubHandler, mockGoogleHandler) {
				t.Error("Expected handler method to be called but it wasn't")
			}

			// Check status code
			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}
}
