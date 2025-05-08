package login

import (
	"black-lotus/internal/domain/models"
	"context"
)

// Repository defines database operations needed by login
type Repository interface {
	// Get user by email to check if user exists
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// Authenticate user with email and password
	LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error)
}
