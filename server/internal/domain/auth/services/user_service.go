package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	authRepositories "black-lotus/internal/domain/auth/repositories"
	tripRepositories "black-lotus/internal/domain/trip/repositories"
	"black-lotus/internal/models"
)

type UserService struct {
	userRepo authRepositories.UserRepositoryInterface
	tripRepo tripRepositories.TripRepositoryInterface
}

type UserServiceInterface interface {
	CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error)
	LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error)
}

func NewUserService(userRepo authRepositories.UserRepositoryInterface, tripRepo tripRepositories.TripRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo, tripRepo: tripRepo}
}

func (s *UserService) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	// Check if user already exists with this email
	existingUser, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password if provided
	var hashedPassword *string
	if input.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		hashStr := string(hash)
		hashedPassword = &hashStr
	}

	// Create user
	user, err := s.userRepo.CreateUser(ctx, input, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Remove sensitive data before returning
	user.HashedPassword = nil

	return user, nil
}

// LoginUser authenticates a user and returns the user if successful
func (s *UserService) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	// Get user by email and password
	user, err := s.userRepo.LoginUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// You could add additional checks here if needed
	// For example, check if email is verified
	if !user.EmailVerified {
		// Decide whether to return an error or just a warning
		// For now, we'll still allow login but you might want to change this
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Don't return the hashed password
	user.HashedPassword = nil

	return user, nil
}

/*
GetUserWithTrips fetches a user and their trips
With primary focus being on Trip domain - i.e. trip summary dashboard would require minimal user info
*/
func (s *UserService) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.User, error) {
	// Get the user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
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
