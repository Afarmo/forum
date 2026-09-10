package models

import "time"

type User struct {
	Id             int       `json:"id"`
	UserName       string    `json:"name"`
	Email          string    `json:"email"`
	Password       string    `json:"-"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}
