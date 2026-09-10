package models

import "time"

type Session struct {
	Id        int       `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	UserId    int       `json:"user_id"`
}
