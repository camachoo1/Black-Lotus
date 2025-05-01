package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"black-lotus/internal/trip/models"
	"black-lotus/internal/trip/repositories"
)

type TripService struct {
	tripRepo *repositories.TripRepository
}

func NewTripService(tripRepo *repositories.TripRepository) *TripService {
	return &TripService{tripRepo: tripRepo}
}

func (s *TripService) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	// Validate dates from user
	if input.EndDate.Before(input.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	// If name is empty, we generate a default name for the Trip
	if input.Name == "" {
		input.Name = fmt.Sprintf("Trip to %s", input.Destination)
	}

	// Create the trip in the DB
	trip, err := s.tripRepo.CreateTrip(ctx, userID, input)

	if err != nil {
		return nil, err
	}

	return trip, nil
}

// GetTripByID retrieves a trip by ID, with ownership verification
func (s *TripService) GetTripByID(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (*models.Trip, error) {
	trip, err := s.tripRepo.GetTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if trip.UserID != userID {
		return nil, errors.New("unauthorized access to trip")
	}

	return trip, nil
}
