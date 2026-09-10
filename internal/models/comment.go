package models

import "time"

type Comment struct {
	Id        int       `json:"id"`
	UserId    string    `json:"user_id"`
	PostId    string    `json:"post_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
