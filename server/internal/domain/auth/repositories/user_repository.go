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

/*
IMPLEMENTED FOR TESTING PURPOSES
*/
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error)
	LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error)
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, input models.CreateUserInput, hashedPassword *string) (*models.User, error) {
	user := new(models.User)

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

	return user, nil
}

// LoginUser verifies credentials and returns the user if valid
func (r *UserRepository) LoginUser(ctx context.Context, input models.LoginUserInput) (*models.User, error) {
	user := new(models.User)
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
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := new(models.User)

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

	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)

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

	return user, nil
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

// GetUserWithTrips retrieves a user and their trips in a single operation
func (r *UserRepository) GetUserWithTrips(ctx context.Context, userID uuid.UUID, limit int, offset int) (*models.User, error) {
	// First get the user
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Set a reasonable default for limit
	if limit <= 0 {
		limit = 10
	}

	// Then get their trips
	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, name, description, start_date, end_date, destination, created_at, updated_at
        FROM trips
        WHERE user_id = $1
        ORDER BY start_date DESC
        LIMIT $2 OFFSET $3
    `, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trips := []*models.Trip{}
	for rows.Next() {
		trip := new(models.Trip)
		err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.Name,
			&trip.Description,
			&trip.StartDate,
			&trip.EndDate,
			&trip.Destination,
			&trip.CreatedAt,
			&trip.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	user.Trips = trips
	return user, nil
}
