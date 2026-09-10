package models

import "time"

type Session struct {
	ID        int       `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    int       `json:"user_id"`
}
