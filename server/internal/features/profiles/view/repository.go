package view

import (
	"black-lotus/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

// Repository defines database operations needed by the profile view feature
type Repository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}
