package trips

import (
	"black-lotus/internal/domain/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	userRepo UserRepository
	tripRepo TripRepository
}
type ServiceInterface interface {
	GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error)
}

func NewService(userRepo UserRepository, tripRepo TripRepository) *Service {
	return &Service{userRepo: userRepo, tripRepo: tripRepo}
}

func (s *Service) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error) {
	// Get the user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is nil before proceeding
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Don't return the hashed password
	user.HashedPassword = nil

	// Get the user's trips
	trips, err := s.tripRepo.GetTripsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Attach trips to user
	user.Trips = trips
	return user, nil
}
