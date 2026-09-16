package models

import "time"

type Post struct {
	Content        string    `json:"content"`
	PostID         int       `json:"post_id"`
	UserID         int       `json:"user_id"`
	PictureContent string    `json:"picture_content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type UpdatePost struct {
	Content        *string `json:"content"`
	PictureContent *string  `json:"picture_content"`
}
