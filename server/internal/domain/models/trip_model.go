package models

import (
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	// Will generate default names for Trips in service file
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	Location    string    `json:"location" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        *User     `json:"-,omitempty"`
}

type CreateTripInput struct {
	// Will generate default names for Trips in service file
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	Location    string    `json:"location" validate:"required"`
}

type UpdateTripInput struct {
	Name        *string    `json:"name" validate:"omitempty,min=1"`
	Description *string    `json:"description"`
	StartDate   *time.Time `json:"start_date" validate:"omitempty"`
	EndDate     *time.Time `json:"end_date" validate:"omitempty"`
	Location    *string    `json:"location" validate:"omitempty,min=1"`
}
