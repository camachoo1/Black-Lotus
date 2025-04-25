package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"black-lotus/internal/services"
)

// AuthMiddleware provides authentication and authorization for routes
type AuthMiddleware struct {
  sessionService *services.SessionService
  userService    *services.UserService
}

// NewAuthMiddleware creates a middleware instance with the required services
func NewAuthMiddleware(sessionService *services.SessionService, userService *services.UserService) *AuthMiddleware {
  return &AuthMiddleware{
    sessionService: sessionService,
    userService:    userService,
  }
}

// Authenticate checks for a valid session before allowing access to protected routes
// Used as an Echo middleware on route groups that require authentication
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
  return func(c echo.Context) error {
        // Extract session cookie
        cookie, err := c.Cookie("session_token")
        if err != nil {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "You must be logged in to access this resource",
            })
        }
        
        // Get the token value
        tokenValue := cookie.Value
        
        // Fetch session by token, not by ID
        session, err := m.sessionService.ValidateSessionByToken(c.Request().Context(), tokenValue)
        if err != nil {
            // Clear invalid cookie
            cookie := new(http.Cookie)
            cookie.Name = "session_token"
            cookie.Value = ""
            cookie.MaxAge = -1
            cookie.Path = "/"
            c.SetCookie(cookie)
            
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Session expired or invalid",
            })
        }
        
        // Fetch user
        user, err := m.userService.GetUserByID(c.Request().Context(), session.UserID)
    
    // Add user to request context for handlers to access
    c.Set("user", user)
    
    return next(c)
  }
}