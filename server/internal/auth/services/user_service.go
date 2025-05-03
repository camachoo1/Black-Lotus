package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"black-lotus/internal/auth/models"
	"black-lotus/internal/auth/repositories"
)

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
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
