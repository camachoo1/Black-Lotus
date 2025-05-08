package register

import (
	"black-lotus/internal/domain/models"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	// Check if user already exists with this email
	existingUser, err := s.repo.GetUserByEmail(ctx, input.Email)
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
	user, err := s.repo.CreateUser(ctx, input, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Remove sensitive data before returning
	user.HashedPassword = nil

	return user, nil
}
