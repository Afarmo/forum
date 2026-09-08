package models

type PostReaction struct {
	PostId   int `json:"post_id"`
	UserId   int `json:"user_id"`
	Reaction int `json:"reaction"`
}

type CommentReaction struct {
	PostId   int `json:"post_id"`
	UserId   int `json:"user_id"`
	Reaction int `json:"reaction"`
}
