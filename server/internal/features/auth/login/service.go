package login

import (
	"black-lotus/internal/domain/models"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	// Get user by email and password
	user, err := s.repo.LoginUser(ctx, input)
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
