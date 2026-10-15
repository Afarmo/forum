package models

import (
	"strings"
	"time"
)

type Post struct {
	Content        string    `json:"content"`
	Title          string    `json:"title"`
	PostID         int       `json:"post_id"`
	UserID         int       `json:"user_id"`
	PictureContent string    `json:"picture_content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Author details, filled in when a single post is loaded.
	AuthorName    string `json:"author_name,omitempty"`
	AuthorPicture string `json:"author_picture,omitempty"`
}

// ImageSrc is the <img src> for the post's picture: a linked image's full URL
// as-is, or the path of an uploaded file under /static/img/post/.
func (p Post) ImageSrc() string {
	if strings.HasPrefix(p.PictureContent, "http://") || strings.HasPrefix(p.PictureContent, "https://") {
		return p.PictureContent
	}
	return "/static/img/post/" + p.PictureContent
}

type UpdatePost struct {
	Title          *string `json:"title"`
	Content        *string `json:"content"`
	PictureContent *string `json:"picture_content"`
}
