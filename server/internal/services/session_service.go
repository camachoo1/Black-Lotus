package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"black-lotus/internal/models"
	"black-lotus/internal/repositories"
)

// SessionService coordinates business logic for session management
type SessionService struct {
	sessionRepo *repositories.SessionRepository
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo *repositories.SessionRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

// CreateSession creates a new session for a user with default duration
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	// Default session duration - 7 days
	sessionDuration := 7 * 24 * time.Hour
	
	return s.sessionRepo.CreateSession(ctx, userID, sessionDuration)
}

// ValidateSession checks if a session exists and is valid
func (s *SessionService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	return s.sessionRepo.GetSessionByID(ctx, sessionID)
}

func (s *SessionService) ValidateSessionByToken(ctx context.Context, token string) (*models.Session, error) {
    return s.sessionRepo.GetSessionByToken(ctx, token)
}

// EndSession terminates a specific session
func (s *SessionService) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.DeleteSession(ctx, sessionID)
}

// EndAllUserSessions terminates all sessions for a user
func (s *SessionService) EndAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteUserSessions(ctx, userID)
}