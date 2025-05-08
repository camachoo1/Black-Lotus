package register

import (
	"black-lotus/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

// Repository defines database operations needed by registration
type Repository interface {
	// Check if email already exists
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// Create a new user
	CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error)

	// Mark email as verified
	SetEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error
}
