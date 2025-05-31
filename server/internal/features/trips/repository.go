package trips

import (
	"context"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
)

type Repository interface {
	CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
	UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	DeleteTrip(ctx context.Context, tripID uuid.UUID) error
	GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
	GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
}
