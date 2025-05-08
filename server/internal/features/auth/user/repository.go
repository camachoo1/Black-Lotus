package user

import (
	"black-lotus/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

// Repository defines database operations needed by the user feature
type Repository interface {
	// Get user by ID
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}
