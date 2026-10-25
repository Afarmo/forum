package models

import (
	"strconv"
	"strings"
	"time"
)

type Post struct {
	Content        string     `json:"content"`
	Title          string     `json:"title"`
	PostID         int        `json:"post_id"`
	UserID         int        `json:"user_id"`
	PictureContent string     `json:"picture_content"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	// Author and category details, filled in when posts are loaded for display.
	AuthorName    string `json:"author_name,omitempty"`
	AuthorPicture string `json:"author_picture,omitempty"`
	CategoryNames string `json:"category_names,omitempty"`
}

// ImageSrc is the <img src> for the post's picture: a linked image's full URL
// as-is, or the path of an uploaded file under /static/img/post/.
func (p Post) ImageSrc() string {
	if strings.HasPrefix(p.PictureContent, "http://") || strings.HasPrefix(p.PictureContent, "https://") {
		return p.PictureContent
	}
	return "/static/img/post/" + p.PictureContent
}

func (p Post) Preview() string {
	content := strings.ReplaceAll(strings.ReplaceAll(p.Content, "*", ""), "\n", " ")
	runes := []rune(content)
	if len(runes) > 210 {
		return string(runes[:210]) + "…"
	}
	return content
}
func (p Post) IsDeleted() bool {
	return p.DeletedAt != nil
}
func (p Post) TimeAgo() string {
	elapsed := time.Since(p.CreatedAt)
	if elapsed < time.Minute {
		return "just now"
	}
	if elapsed < time.Hour {
		return strconv.Itoa(int(elapsed/time.Minute)) + "m ago"
	}
	if elapsed < 24*time.Hour {
		return strconv.Itoa(int(elapsed/time.Hour)) + "h ago"
	}
	if elapsed < 8*24*time.Hour {
		return strconv.Itoa(int(elapsed/(24*time.Hour))) + "d ago"
	}
	return p.CreatedAt.Format("Jan 2")
}

type UpdatePost struct {
	Title          *string `json:"title"`
	Content        *string `json:"content"`
	PictureContent *string `json:"picture_content"`
}
