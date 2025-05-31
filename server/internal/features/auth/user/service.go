package user

import (
	"black-lotus/internal/domain/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ServiceInterface interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is nil before accessing properties
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Remove sensitive information before returning
	if user.HashedPassword != nil {
		user.HashedPassword = nil
	}
	return user, nil
}
