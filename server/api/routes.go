package api

import (
	"github.com/labstack/echo/v4"
	
	"black-lotus/api/controllers"
	"black-lotus/db"
	"black-lotus/internal/repositories"
	"black-lotus/internal/services"
)

func RegisterRoutes(e *echo.Echo) {
	// Create repositories
	userRepo := repositories.NewUserRepository(db.DB)
	
	// Create services
	userService := services.NewUserService(userRepo)
	
	// Create controllers
	userController := controllers.NewUserController(userService)
	
	// Public routes
	e.POST("/api/users", userController.RegisterUser)
	
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})
}