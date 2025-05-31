package view

import (
	"black-lotus/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

type ServiceInterface interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is nil before accessing properties
	if user == nil {
		return nil, nil
	}

	// Don't return the hashed password
	user.HashedPassword = nil
	return user, nil
}
