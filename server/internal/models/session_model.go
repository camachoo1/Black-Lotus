package models

import (
	"time"

	"github.com/google/uuid"

)

type Session struct {
    ID        uuid.UUID `json:"id"`
    UserID    uuid.UUID `json:"user_id"`
    ExpiresAt time.Time `json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
}