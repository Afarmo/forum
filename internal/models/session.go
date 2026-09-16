package models

import (
	"time"
	"uuid"
)

type Session struct {
	ID        uuid.UUID `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    int       `json:"user_id"`
}
