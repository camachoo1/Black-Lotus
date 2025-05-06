package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	authRepository "black-lotus/internal/domain/auth/repositories"
	tripRepository "black-lotus/internal/domain/trip/repositories"
	"black-lotus/internal/models"
)

/*
*
IMPLEMENTED FOR TESTING PURPOSES
*/
type TripServiceInterface interface {
	CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	UpdateTrip(ctx context.Context, tripID uuid.UUID, userID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	DeleteTrip(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) error
	GetTripByID(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (*models.Trip, error)
	GetTripWithUser(ctx context.Context, tripID uuid.UUID, requestUserID uuid.UUID) (*models.Trip, error)
	GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error)
	GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error)
}
type TripService struct {
	tripRepo tripRepository.TripRepositoryInterface
	userRepo authRepository.UserRepositoryInterface
}

func NewTripService(tripRepo tripRepository.TripRepositoryInterface, userRepo authRepository.UserRepositoryInterface) *TripService {
	return &TripService{tripRepo: tripRepo, userRepo: userRepo}
}

func (s *TripService) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	// Validate dates from user
	if input.EndDate.Before(input.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	// If name is empty, we generate a default name for the Trip
	if input.Name == "" {
		input.Name = fmt.Sprintf("Trip to %s", input.Location)
	}

	// Create the trip in the DB
	trip, err := s.tripRepo.CreateTrip(ctx, userID, input)

	if err != nil {
		return nil, err
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

// DeleteTrip deletes a trip with ownership verification
func (s *TripService) DeleteTrip(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) error {
	// Verify ownership of the trip
	trip, err := s.tripRepo.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip.UserID != userID {
		return errors.New("unauthorized access to trip")
	}

	return s.tripRepo.DeleteTrip(ctx, tripID)
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

func (s *TripService) GetTripWithUser(ctx context.Context, tripID uuid.UUID, requestUserID uuid.UUID) (*models.Trip, error) {
	trip, err := s.tripRepo.GetTripWithUser(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Verify the requesting user has permission to see this trip
	if trip.UserID != requestUserID {
		return nil, errors.New("unauthorized access to trip")
	}

	return trip, nil
}

func (s *TripService) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error) {
	// First, verify the user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Then get their trips
	trips, err := s.tripRepo.GetTripsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Attach trips to user
	user.Trips = trips
	return user, nil
}

func (s *TripService) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error) {
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	trips, err := s.tripRepo.GetTripsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return trips, nil
}
