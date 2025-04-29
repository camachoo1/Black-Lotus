package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
    ID        uuid.UUID `json:"id"`
    UserID    uuid.UUID `json:"user_id"`
    AccessToken  string    `json:"-"` // Short-lived token
    RefreshToken string    `json:"-"` // Long-lived token
    AccessExpiry time.Time `json:"access_expires_at"`
    RefreshExpiry time.Time `json:"refresh_expires_at"`
    CreatedAt time.Time `json:"created_at"`
}