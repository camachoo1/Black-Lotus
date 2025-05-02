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

// UpdateTrip updates a trip with ownership verification
func (s *TripService) UpdateTrip(ctx context.Context, tripID uuid.UUID, userID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	// First, verify ownership
	trip, err := s.tripRepo.GetTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	if trip.UserID != userID {
		return nil, errors.New("unauthorized access to trip")
	}

	// If updating dates, validate them
	if input.StartDate != nil && input.EndDate != nil {
		if input.EndDate.Before(*input.StartDate) {
			return nil, errors.New("end date cannot be before start date")
		}
	} else if input.StartDate != nil && trip.EndDate.Before(*input.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	} else if input.EndDate != nil && input.EndDate.Before(trip.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	// TODO: validate location if it's being updated!!

	// Update the trip
	return s.tripRepo.UpdateTrip(ctx, tripID, input)
}
