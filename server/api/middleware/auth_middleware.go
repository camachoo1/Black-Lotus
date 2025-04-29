package middleware

import (
	"black-lotus/internal/services"

	"net/http"

	"github.com/labstack/echo/v4"
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

// Authenticate checks for a valid access token before allowing access to protected routes
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Extract access token cookie
        accessCookie, err := c.Cookie("access_token")
        if err != nil {
            // No access token - check if there's a refresh token
            _, refreshErr := c.Cookie("refresh_token")
            if refreshErr != nil {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "You must be logged in to access this resource",
                })
            }
            
            // Has refresh token but no access token
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Access token expired",
                "code": "token_expired",
            })
        }
        
        // Validate access token
        session, err := m.sessionService.ValidateAccessToken(c.Request().Context(), accessCookie.Value)
        if err != nil {
            // Clear invalid access token cookie
            expiredCookie := new(http.Cookie)
            expiredCookie.Name = "access_token"
            expiredCookie.Value = ""
            expiredCookie.MaxAge = -1
            expiredCookie.Path = "/"
            c.SetCookie(expiredCookie)
            
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Access token expired or invalid",
                "code": "token_invalid",
            })
        }
        
        // Fetch user
        user, err := m.userService.GetUserByID(c.Request().Context(), session.UserID)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to get user information",
            })
        }
        
        // Add user to request context for handlers to access
        c.Set("user", user)
        return next(c)
    }
}