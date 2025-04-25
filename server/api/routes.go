package api

import (
	"github.com/labstack/echo/v4"

	"black-lotus/api/controllers"
	"black-lotus/api/middleware"
	"black-lotus/db"
	"black-lotus/internal/repositories"
	"black-lotus/internal/services"
)

func AuthRoutes(e *echo.Echo) {
	// Create repositories
	userRepo := repositories.NewUserRepository(db.DB)
	sessionRepo := repositories.NewSessionRepository(db.DB)
	
	// Create services
	userService := services.NewUserService(userRepo)
	sessionService := services.NewSessionService(sessionRepo) 
	
	// Create controllers
	userController := controllers.NewUserController(userService, sessionService)

	// Create auth middlewares
	authMiddleware := middleware.NewAuthMiddleware(sessionService, userService)
	
	// Public Routes
	e.POST("/api/signup", userController.RegisterUser)
	e.POST("/api/login", userController.LoginUser)
	e.POST("/api/logout", userController.LogoutUser)

	// Private Routes
	protected := e.Group("/api")
	protected.Use(authMiddleware.Authenticate)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})
}