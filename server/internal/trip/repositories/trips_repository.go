package repositories

import (
	"context"
	"errors"

	"black-lotus/internal/trip/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository struct {
	db *pgxpool.Pool
}

func NewTripRepository(db *pgxpool.Pool) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) CreateTrip(ctx context.Context, userID uuid.UUID, input models.CreateTripInput) (*models.Trip, error) {
    trip := new(models.Trip)
    
    err := r.db.QueryRow(ctx, `
        INSERT INTO trips (user_id, name, description, start_date, end_date, destination)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, user_id, name, description, start_date, end_date, destination, created_at, updated_at
    `, 
    userID,
    input.Name, 
    input.Description, 
    input.StartDate, 
    input.EndDate, 
    input.Destination).Scan(
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
    
    return trip, nil
}

func (r *TripRepository) GetTripByID(ctx context.Context, tripID uuid.UUID) (*models.Trip, error) {
    trip := new(models.Trip)

    err := r.db.QueryRow(ctx, `
        SELECT id, user_id, name, description, start_date, end_date, destination, created_at, updated_at
        FROM trips
        WHERE id = $1
    `, tripID).Scan(
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
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("trip not found")
        }
        return nil, err
    }
    
    return trip, nil
}
