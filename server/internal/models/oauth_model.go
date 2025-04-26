package models

import (
	"time"

	"github.com/google/uuid"
)

// OAuthAccount model is for a user's connection to an external OAuth provider (i.e. Google, GitHub, etc)
type OAuthAccount struct {
	ProviderID     string    `json:"provider_id"`
	ProviderUserID string    `json:"provider_user_id"`
	UserID         uuid.UUID `json:"user_id"`
	AccessToken    string    `json:"-"` // Not included in JSON responses
	RefreshToken   string    `json:"-"` // Not included in JSON responses
	ExpiresAt      time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}