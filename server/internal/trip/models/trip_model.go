package models

import (
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	// Will generate default names for Trips in service file
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	Destination string    `json:"destination" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// WILL NEED TO ADD RELATIONS AS FEATURES GET IMPLEMENTED
}

type CreateTripInput struct {
	// Will generate default names for Trips in service file
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	Destination string    `json:"destination" validate:"required"`
}

type UpdateTripInput struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Destination *string    `json:"destination"`
}
