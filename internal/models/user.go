package models

import "time"

type User struct {
	ID             int       `json:"id"`
	UserName       string    `json:"name"`
	Email          string    `json:"email"`
	PasswordHash   []byte    `json:"-"`
	PasswordSalt   []byte    `json:"-"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}
