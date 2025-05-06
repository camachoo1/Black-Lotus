package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword *string   `json:"hashed_password,omitempty"`
	EmailVerified  bool      `json:"email_verified" default:"false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Trips          []*Trip   `json:"trips,omitempty"`
}

type CreateUserInput struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Password *string `json:"password" validate:"required,min=6,containsuppercase,containslowercase,containsnumber,containsspecialchar"`
}

type LoginUserInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
