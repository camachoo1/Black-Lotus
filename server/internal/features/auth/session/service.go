package session

import (
	"context"

	"github.com/google/uuid"

	"black-lotus/internal/domain/models"
)

type Service struct {
	repo Repository
}

type ServiceInterface interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error)
	ValidateAccessToken(ctx context.Context, token string) (*models.Session, error)
	ValidateRefreshToken(ctx context.Context, token string) (*models.Session, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*models.Session, error)
	EndSessionByAccessToken(ctx context.Context, token string) error
	EndSessionByRefreshToken(ctx context.Context, token string) error
	EndAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

func NewService(repo Repository) ServiceInterface {
	return &Service{repo: repo}
}

func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	return s.repo.CreateSession(ctx, userID, AccessTokenDuration, RefreshTokenDuration)
}

func (s *Service) ValidateAccessToken(ctx context.Context, token string) (*models.Session, error) {
	return s.repo.GetSessionByAccessToken(ctx, token)
}

func (s *Service) ValidateRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	return s.repo.GetSessionByRefreshToken(ctx, token)
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	// First validate the refresh token
	session, err := s.repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Then get a new access token
	return s.repo.RefreshAccessToken(ctx, session.ID)
}

func (s *Service) EndSessionByAccessToken(ctx context.Context, token string) error {
	return s.repo.DeleteSessionByAccessToken(ctx, token)
}

func (s *Service) EndSessionByRefreshToken(ctx context.Context, token string) error {
	return s.repo.DeleteSessionByRefreshToken(ctx, token)
}

func (s *Service) EndAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteUserSessions(ctx, userID)
}
