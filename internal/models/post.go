package models

import "time"

type Post struct {
	Content        string    `json:"content"`
	PostId         int       `json:"post_id"`
	UserId         int       `json:"user_id"`
	PictureContent string    `json:"picture_content"`
	CreatedAt      time.Time `json:"created_at"`
}
