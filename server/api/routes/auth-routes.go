package routes

import (
	"github.com/labstack/echo/v4"

	controllers "black-lotus/api/controllers/auth"
	"black-lotus/api/middleware"
	"black-lotus/db"
	"black-lotus/internal/auth/repositories"
	"black-lotus/internal/auth/services"
)

func AuthRoutes(e *echo.Echo) {
	// Create repositories
	userRepo := repositories.NewUserRepository(db.DB)
	sessionRepo := repositories.NewSessionRepository(db.DB)
	oauthRepo := repositories.NewOAuthRepository((db.DB))

	// Create services
	userService := services.NewUserService(userRepo)
	sessionService := services.NewSessionService(sessionRepo)
	oauthService := services.NewOAuthService(oauthRepo, userRepo)

	// Create controllers
	userController := controllers.NewUserController(userService, sessionService)
	oauthController := controllers.NewOAuthController(oauthService, sessionService)

	// Create auth middlewares
	authMiddleware := middleware.NewAuthMiddleware(sessionService, userService)

	// Public Routes
	e.POST("/api/signup", userController.RegisterUser)
	e.POST("/api/login", userController.LoginUser)
	e.POST("/api/logout", userController.LogoutUser)
	e.GET("/api/csrf-token", userController.GetCSRFToken)

	// OAuth Routes
	e.GET("/api/auth/github", oauthController.GetGitHubAuthURL)
	e.GET("/api/auth/github/callback", oauthController.HandleGitHubCallback)
	e.GET("/api/auth/google", oauthController.GetGoogleAuthURL)
	e.GET("/api/auth/google/callback", oauthController.HandleGoogleCallback)

	// Private Routes
	protected := e.Group("/api")
	protected.Use(authMiddleware.Authenticate)
	protected.GET("/profile", userController.GetUserProfile)
}
