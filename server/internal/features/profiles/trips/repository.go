package trips

import (
	"black-lotus/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

// UserRepository defines user operations needed by the trips feature
type UserRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

// TripRepository defines trip operations needed by the trips feature
type TripRepository interface {
	GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error)
}
