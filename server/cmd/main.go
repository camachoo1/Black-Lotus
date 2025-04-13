package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
  // Echo instance
  e := echo.New()

  // Middleware
  e.Use(middleware.Logger())
  e.Use(middleware.Recover())

  // Routes
  e.GET("/", welcome)

  // Start server
  if err := e.Start(":5001"); err != nil && !errors.Is(err, http.ErrServerClosed) {
    slog.Error("failed to start server", "error", err)
  }
}

// Handler
func welcome(c echo.Context) error {
  return c.String(http.StatusOK, "Welcome to Black Lotus!")
}