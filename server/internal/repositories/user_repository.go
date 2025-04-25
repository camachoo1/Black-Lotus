package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"black-lotus/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
    var user models.User
    
    err := r.db.QueryRow(ctx, `
        INSERT INTO users (name, email, hashed_password)
        VALUES ($1, $2, $3)
        RETURNING id, name, email, hashed_password, email_verified, created_at, updated_at
    `, input.Name, input.Email, hashedPassword).Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &user.HashedPassword,
        &user.EmailVerified,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, hashed_password, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}
	
	return &user, nil
}