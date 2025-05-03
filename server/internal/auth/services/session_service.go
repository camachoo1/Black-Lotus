package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/auth/models"
	"black-lotus/internal/auth/repositories"
)

// SessionService coordinates business logic for session management
type SessionService struct {
	sessionRepo repositories.SessionRepositoryInterface
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo repositories.SessionRepositoryInterface) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

// CreateSession creates a new session with both access and refresh tokens
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	// Access token - short-lived (1 hour)
	accessDuration := 1 * time.Hour
	// Refresh token - long-lived (7 days)
	refreshDuration := 7 * 24 * time.Hour

	return s.sessionRepo.CreateSession(ctx, userID, accessDuration, refreshDuration)
}

// ValidateAccessToken checks if an access token is valid
func (s *SessionService) ValidateAccessToken(ctx context.Context, token string) (*models.Session, error) {
	return s.sessionRepo.GetSessionByAccessToken(ctx, token)
}

// ValidateRefreshToken checks if a refresh token is valid
func (s *SessionService) ValidateRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	return s.sessionRepo.GetSessionByRefreshToken(ctx, token)
}

// RefreshAccessToken uses a refresh token to generate a new access token
func (s *SessionService) RefreshAccessToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	// First validate the refresh token
	session, err := s.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate a new access token
	return s.sessionRepo.RefreshAccessToken(ctx, session.ID)
}

// EndSessionByAccessToken terminates a session by its access token
func (s *SessionService) EndSessionByAccessToken(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteSessionByAccessToken(ctx, token)
}

// EndSessionByRefreshToken terminates a session by its refresh token
func (s *SessionService) EndSessionByRefreshToken(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteSessionByRefreshToken(ctx, token)
}

// EndAllUserSessions terminates all sessions for a user
func (s *SessionService) EndAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteUserSessions(ctx, userID)
}
