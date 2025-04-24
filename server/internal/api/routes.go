package api

import (
	"context"
	"net/http"

	"black-lotus/internal/db"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes(e *echo.Echo) {
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Add this to your RegisterRoutes function in api/routes.go
	e.POST("/users/test", func(c echo.Context) error {
    // Create a test user
    var userId string
    err := db.DB.QueryRow(context.Background(), `
        INSERT INTO users (name, email, email_verified) 
        VALUES ('Test User', 'test@example.com', false) 
        RETURNING id
    `).Scan(&userId)
    
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "status": "error",
            "message": "Failed to create test user: " + err.Error(),
        })
    }

    // Create a test session for this user
    var sessionId string
    err = db.DB.QueryRow(context.Background(), `
        INSERT INTO sessions (user_id, expires_at) 
        VALUES ($1, NOW() + INTERVAL '7 days') 
        RETURNING id
    `, userId).Scan(&sessionId)

    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "status": "error",
            "message": "Failed to create test session: " + err.Error(),
        })
    }

    // Create a verification code for this user
    var verificationId string
    err = db.DB.QueryRow(context.Background(), `
        INSERT INTO email_verifications (code, user_id) 
        VALUES ('123456', $1) 
        RETURNING id
    `, userId).Scan(&verificationId)

    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "status": "error",
            "message": "Failed to create verification code: " + err.Error(),
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "status": "success",
        "message": "Test user created successfully with session and verification",
        "userId": userId,
        "sessionId": sessionId,
        "verificationId": verificationId,
    })
	})
	// Add your other routes here
	// e.GET("/users", GetUsers)
	// e.POST("/users", CreateUser)
	// etc.
}