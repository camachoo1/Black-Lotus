package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

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

// LoginUser verifies credentials and returns the user if valid
func (r *UserRepository) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
    var user models.User
    var hashedPassword string
    
    // Retrieve user and hashed password from database
    err := r.db.QueryRow(ctx, `
        SELECT id, name, email, hashed_password, email_verified, created_at
        FROM users
        WHERE email = $1 AND hashed_password IS NOT NULL
    `, input.Email).Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &hashedPassword,
        &user.EmailVerified,
        &user.CreatedAt,
    )
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("invalid email or password")
        }
        return nil, err
    }
    
    // Verify password
    err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
    if err != nil {
        return nil, errors.New("invalid email or password")
    }
    
    // Success - don't include password hash in returned user
    return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
    var user models.User
    
    err := r.db.QueryRow(ctx, `
        SELECT id, name, email, hashed_password, email_verified, created_at, updated_at
        FROM users
        WHERE id = $1
    `, userID).Scan(
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

// Changing verified email to true - used for oauth (will implement verification email later)
func (r *UserRepository) SetEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users 
		SET email_verified = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, verified, userID)
	
	return err
}