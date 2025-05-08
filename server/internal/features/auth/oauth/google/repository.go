package google

import (
	"black-lotus/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

// UserRepository defines user operations needed by Google OAuth
type UserRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error)
	SetEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error
}

// OAuthRepository defines OAuth account operations
type OAuthRepository interface {
	CreateOAuthAccount(ctx context.Context, account models.OAuthAccount) error
	GetOAuthAccount(ctx context.Context, providerID, providerUserID string) (*models.OAuthAccount, error)
}
