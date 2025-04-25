package api

import (
	"github.com/labstack/echo/v4"
	
	"black-lotus/api/controllers"
	"black-lotus/db"
	"black-lotus/internal/repositories"
	"black-lotus/internal/services"
)

func AuthRoutes(e *echo.Echo) {
	// Create repositories
	userRepo := repositories.NewUserRepository(db.DB)
	
	// Create services
	userService := services.NewUserService(userRepo)
	
	// Create controllers
	userController := controllers.NewUserController(userService)
	
	// Public routes
	e.POST("/api/signup", userController.RegisterUser)
	e.POST("/api/login", userController.LoginUser)
	
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})
}