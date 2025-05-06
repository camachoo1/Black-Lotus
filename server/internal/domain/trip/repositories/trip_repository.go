package repositories

import (
	"context"
	"errors"

	"black-lotus/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository struct {
	db *pgxpool.Pool
}

/*
IMPLEMENTED FOR TESTING PURPOSES
*/
type TripRepositoryInterface interface {
	CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error)
	GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
	UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error)
	DeleteTrip(ctx context.Context, tripID uuid.UUID) error
	GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Trip, error)
	GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error)
}

func NewTripRepository(db *pgxpool.Pool) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
	trip := new(models.Trip)

	err := r.db.QueryRow(ctx, `
        INSERT INTO trips (user_id, name, description, start_date, end_date, location)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, user_id, name, description, start_date, end_date, location, created_at, updated_at
    `,
		userID,
		input.Name,
		input.Description,
		input.StartDate,
		input.EndDate,
		input.Location).Scan(
		&trip.ID,
		&trip.UserID,
		&trip.Name,
		&trip.Description,
		&trip.StartDate,
		&trip.EndDate,
		&trip.Location,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return trip, nil
}

// UpdateTrip updates an existing trip
func (r *TripRepository) UpdateTrip(ctx context.Context, tripID uuid.UUID, input models.UpdateTripInput) (*models.Trip, error) {
	trip := new(models.Trip)

	err := r.db.QueryRow(ctx, `
	UPDATE trips
	SET 
	name = COALESCE($1, name),
	description = COALESCE($2, description),
	start_date = COALESCE($3, start_date),
	end_date = COALESCE($4, end_date),
	location = COALESCE($5, location),
	updated_at = NOW()
	WHERE id = $6
	RETURNING id, user_id, name, description, start_date, end_date, location, created_at, updated_at
	`,
		input.Name,
		input.Description,
		input.StartDate,
		input.EndDate,
		input.Location,
		tripID).Scan(
		&trip.ID,
		&trip.UserID,
		&trip.Name,
		&trip.Description,
		&trip.StartDate,
		&trip.EndDate,
		&trip.Location,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("trip not found")
		}
		return nil, err
	}

	return trip, nil
}

// DeleteTrip removes trip from DB.
func (r *TripRepository) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	commandTag, err := r.db.Exec(ctx, `
	DELETE FROM trips
	WHERE id = $1
	`, tripID)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("trip not found")
	}

	return nil
}

// GetTripByID returns a specific trip based on ID
func (r *TripRepository) GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	trip := new(models.Trip)

	err := r.db.QueryRow(ctx, `
				SELECT id, user_id, name, description, start_date, end_date, location, created_at, updated_at
				FROM trips
				WHERE id = $1
		`, tripID).Scan(
		&trip.ID,
		&trip.UserID,
		&trip.Name,
		&trip.Description,
		&trip.StartDate,
		&trip.EndDate,
		&trip.Location,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("trip not found")
		}
		return nil, err
	}

	return trip, nil
}

// GetTripsByUserID fetches all trips for a given user.
func (r *TripRepository) GetTripsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Trip, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, name, description, start_date, end_date, location, created_at, updated_at
        FROM trips
        WHERE user_id = $1
        ORDER BY start_date DESC
        LIMIT $2 OFFSET $3
    `, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*models.Trip

	for rows.Next() {
		trip := new(models.Trip)

		err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.Name,
			&trip.Description,
			&trip.StartDate,
			&trip.EndDate,
			&trip.Location,
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

	return trips, nil
}

// GetTripWithUser retrieves a trip and its user in a single operation
func (r *TripRepository) GetTripWithUser(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
	// Get the trip first
	trip, err := r.GetTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Then get the user
	user := new(models.User)
	err = r.db.QueryRow(ctx, `
        SELECT id, name, email, email_verified, created_at, updated_at
        FROM users
        WHERE id = $1
    `, trip.UserID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Don't include password in the user object
	trip.User = user
	return trip, nil
}
