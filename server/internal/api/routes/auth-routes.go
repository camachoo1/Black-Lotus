// server/internal/api/routes/auth_routes.go
package routes

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/common/middleware"
	"black-lotus/internal/features/auth/login"
	"black-lotus/internal/features/auth/oauth"
	"black-lotus/internal/features/auth/oauth/github"
	"black-lotus/internal/features/auth/oauth/google"
	"black-lotus/internal/features/auth/register"
	"black-lotus/internal/features/auth/session"
	"black-lotus/internal/features/auth/user"
	"black-lotus/internal/features/profiles/view"
	"black-lotus/internal/infrastructure/repositories"
	"black-lotus/pkg/db"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(e *echo.Echo, validator *validator.Validate) {
	// Create repositories - these implement all the feature-specific interfaces
	userRepo := repositories.NewUserRepository(db.DB)
	sessionRepo := repositories.NewSessionRepository(db.DB)
	oauthRepo := repositories.NewOAuthRepository(db.DB)

	// Create session service (used by multiple features)
	sessionService := session.NewService(sessionRepo)

	// Create feature-specific services
	loginService := login.NewService(userRepo)
	registerService := register.NewService(userRepo)
	userService := user.NewService(userRepo)
	profileService := view.NewService(userRepo)

	// Create OAuth provider services
	githubService := github.NewService(oauthRepo, userRepo)
	googleService := google.NewService(oauthRepo, userRepo)

	// Create provider-specific handlers
	githubHandler := github.NewHandler(githubService, sessionService)
	googleHandler := google.NewHandler(googleService, sessionService)

	// Create feature-specific handlers
	loginHandler := login.NewHandler(loginService, sessionService, validator)
	registerHandler := register.NewHandler(registerService, sessionService, validator)
	userHandler := user.NewHandler(userService)
	sessionHandler := session.NewHandler(sessionService)
	profileHandler := view.NewHandler(profileService, sessionService)

	// Create OAuth main handler that composes provider handlers
	oauthHandler := oauth.NewHandler(githubHandler, googleHandler)

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionService, userService)

	// Public Routes
	e.POST("/api/signup", registerHandler.Register)
	e.POST("/api/login", loginHandler.Login)
	e.POST("/api/logout", sessionHandler.LogoutUser)
	e.GET("/api/csrf-token", sessionHandler.GetCSRFToken)

	// OAuth Routes
	e.GET("/api/auth/github", oauthHandler.GetGitHubAuthURL)
	e.GET("/api/auth/github/callback", oauthHandler.HandleGitHubCallback)
	e.GET("/api/auth/google", oauthHandler.GetGoogleAuthURL)
	e.GET("/api/auth/google/callback", oauthHandler.HandleGoogleCallback)

	// Private Routes
	protected := e.Group("/api")
	protected.Use(authMiddleware.Authenticate)
	protected.GET("/user/:id", userHandler.GetUserByID)
	protected.GET("/profile", profileHandler.GetUserProfile)
}
