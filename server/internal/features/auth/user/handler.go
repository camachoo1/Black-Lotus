package user

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type HandlerInterface interface {
	GetUserByID(ctx echo.Context) error
}

type Handler struct {
	userService ServiceInterface
}

func NewHandler(userService ServiceInterface) HandlerInterface {
	return &Handler{
		userService: userService,
	}
}

func (h *Handler) GetUserByID(ctx echo.Context) error {
	// Get user ID from path parameter
	idParam := ctx.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID format",
		})
	}

	// Get user by ID
	user, err := h.userService.GetUserByID(ctx.Request().Context(), userID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user: " + err.Error(),
		})
	}

	// Add explicit check for nil user
	if user == nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return ctx.JSON(http.StatusOK, user)
}
