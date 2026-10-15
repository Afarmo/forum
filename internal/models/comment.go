package models

import "time"

type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ParentID  *int      `json:"parent_id"`
	PostID    int       `json:"post_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AuthorName    string    `json:"author_name,omitempty"`
	AuthorPicture string    `json:"author_picture,omitempty"`
	Replies       []Comment `json:"replies,omitempty"`
}
