package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo *echo.Echo
}

func NewServer() *Server {

	// Initialize Echo
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-CSRF-TOKEN"},
		ExposeHeaders:    []string{"Set-Cookie"},
		AllowCredentials: true,  // This is crucial for sending cookies
		MaxAge:           86400, // 1 day to cache preflight requests
	}))
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieHTTPOnly: false,
		CookieMaxAge:   3600, // 1 hour
	}))

	// Rate limiting to prevent abuse
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))) // 20 requests per second

	return &Server{
		echo: e,
	}
}

func (s *Server) Start(port string) error {
	return s.echo.Start(":" + port)
}

func (s *Server) Echo() *echo.Echo {
	return s.echo
}
