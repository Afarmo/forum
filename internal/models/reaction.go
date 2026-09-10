package models

type PostReaction struct {
	PostID   int `json:"post_id"`
	UserID   int `json:"user_id"`
	Reaction int `json:"reaction"`
}

type CommentReaction struct {
	PostID   int `json:"post_id"`
	UserID   int `json:"user_id"`
	Reaction int `json:"reaction"`
}
