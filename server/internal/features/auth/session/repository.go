package session

import (
	"black-lotus/internal/domain/models"
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines database operations needed by session management
type Repository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error)
	GetSessionByAccessToken(ctx context.Context, token string) (*models.Session, error)
	GetSessionByRefreshToken(ctx context.Context, token string) (*models.Session, error)
	RefreshAccessToken(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)
	DeleteSessionByAccessToken(ctx context.Context, token string) error
	DeleteSessionByRefreshToken(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
}
