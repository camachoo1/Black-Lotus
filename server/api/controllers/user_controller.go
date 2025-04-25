package controllers

import (
	"fmt"
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
	validator   *validator.Validate
}
func NewUserController(userService *services.UserService) *UserController {
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
        validator:   validate,
    }
}

func (c *UserController) RegisterUser(ctx echo.Context) error {
	var input models.CreateUserInput
	
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
	
	return ctx.JSON(http.StatusCreated, user)
}

