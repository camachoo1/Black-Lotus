package api

import (
	"github.com/labstack/echo/v4"

	"black-lotus/internal/api/routes"
)

func SetupRouter(e *echo.Echo) *echo.Echo {
	routes.AuthRoutes(e)

	// Test Routes
	e.GET("/oauth-test", func(c echo.Context) error {
		// Set CORS headers to allow access if needed
		c.Response().Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

		return c.File("public/oauth-test.html")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})

	return e
}
