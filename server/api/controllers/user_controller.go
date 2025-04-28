package controllers

import (
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"black-lotus/internal/models"
	"black-lotus/internal/services"
)

type UserController struct {
	userService *services.UserService
    sessionService *services.SessionService
	validator   *validator.Validate
}

func NewUserController(userService *services.UserService, sessionService *services.SessionService) *UserController {
    validate := validator.New()
    
    // This is critical - register struct-level validation
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
    
    return &UserController{
        userService: userService,
        sessionService: sessionService,
        validator:   validate,
    }
}

// Creates a new user account and logs them in automatically
func (c *UserController) RegisterUser(ctx echo.Context) error {
    var input models.CreateUserInput
    
    // Validate request data
    if err := ctx.Bind(&input); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    if err := c.validator.Struct(input); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }
    
    // Create the user
    user, err := c.userService.CreateUser(ctx.Request().Context(), input)
    if err != nil {
        // Check for specific errors
        if err.Error() == "user with this email already exists" {
            return ctx.JSON(http.StatusConflict, map[string]string{
                "error": err.Error(),
            })
        }
        return ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create user",
        })
    }
    
    // Create a session to automatically log in the new user
    session, err := c.sessionService.CreateSession(ctx.Request().Context(), user.ID)
    if err != nil {
        // User was created, but session creation failed
        // We'll still return success but log the error
        log.Printf("Failed to create session for new user: %v", err)
    } else {
        // Set secure cookie with session token
        cookie := new(http.Cookie)
        cookie.Name = "session_token"
        cookie.Value = session.Token
        cookie.Expires = session.ExpiresAt
        cookie.Path = "/"
        cookie.HttpOnly = true  // Prevents JavaScript access
        // WILL ADD BACK IN FOR PRODUCTION

        cookie.Secure = true    // Requires HTTPS
        cookie.SameSite = http.SameSiteStrictMode  // Prevents CSRF attacks
        
        ctx.SetCookie(cookie)
    }
    
    return ctx.JSON(http.StatusCreated, user)
}

// LoginUser authenticates a user and creates a session
func (c *UserController) LoginUser(ctx echo.Context) error {
    var input models.LoginUserInput
    
     // Validate request data
    if err := ctx.Bind(&input); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    if err := c.validator.Struct(input); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }
    
    // Authenticate user credentials
    user, err := c.userService.LoginUser(ctx.Request().Context(), input)
    if err != nil {
        // Generic error for security (don't reveal if email or password was wrong)
        return ctx.JSON(http.StatusUnauthorized, map[string]string{
            "error": "Invalid email or password",
        })
    }
    
    // Create a session for the authenticated user
    session, err := c.sessionService.CreateSession(ctx.Request().Context(), user.ID)
    if err != nil {
        log.Printf("Session creation error: %v", err)
        return ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create session: " + err.Error(),
        })
    }

    // Set secure HTTP-only cookie with session ID
    cookie := new(http.Cookie)
    cookie.Name = "session_token" // Using token instead of ID for better security
    cookie.Value = session.Token
    cookie.Expires = session.ExpiresAt
    cookie.Path = "/"
    cookie.HttpOnly = true // Used to prevent Javascript access

    // WILL ADD BACK IN FOR PRODUCTION
    cookie.Secure = true
    cookie.SameSite = http.SameSiteStrictMode // Prevents CSRF attacks

    ctx.SetCookie(cookie)
    
    return ctx.JSON(http.StatusOK, user)
}

// LogoutUser ends the current user session
func (c *UserController) LogoutUser(ctx echo.Context) error {
	cookie, err := ctx.Cookie("session_token")
	if err != nil {
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Already logged out",
		})
	}
	
	 // Delete the session using the token directly
    err = c.sessionService.EndSessionByToken(ctx.Request().Context(), cookie.Value)
    if err != nil {
        // Log the error but still clear the cookie
        log.Printf("Failed to end session: %v", err)
    }
	
    // Make sure to always clear the cookie, even if session delete fails
	cookie = new(http.Cookie)
	cookie.Name = "session_token"
	cookie.Value = ""
	cookie.MaxAge = -1  // Expire immediately
	cookie.Path = "/"
	ctx.SetCookie(cookie)
	
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

func (c *UserController) GetUserProfile(ctx echo.Context) error {
    // Get session from cookie
    cookie, err := ctx.Cookie("session_token")
    if err != nil {
        return ctx.JSON(http.StatusUnauthorized, map[string]string{
            "error": "Not authenticated",
        })
    }
    
    // Validate session
    session, err := c.sessionService.ValidateSessionByToken(ctx.Request().Context(), cookie.Value)
    if err != nil {
        return ctx.JSON(http.StatusUnauthorized, map[string]string{
            "error": "Invalid session",
        })
    }
    
    // Get user from session
    user, err := c.userService.GetUserByID(ctx.Request().Context(), session.UserID)
    if err != nil {
        return ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to get user",
        })
    }
    
    return ctx.JSON(http.StatusOK, user)
}

func (c *UserController) GetCSRFToken(ctx echo.Context) error {
    token := ctx.Get("csrf").(string)
    
    return ctx.JSON(http.StatusOK, map[string]string{
        "csrf_token": token,
    })
}