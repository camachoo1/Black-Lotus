package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/api/routes"
	validation "black-lotus/internal/common/validations"
)

func SetupRouter(e *echo.Echo) *echo.Echo {
	v := validator.New()
	validation.RegisterPasswordValidators(v)
	routes.RegisterAuthRoutes(e, v)

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
