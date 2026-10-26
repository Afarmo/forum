package service

import (
	"context"
	"time"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type CommentService struct {
	repo *repository.CommentRepository
}

func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{
		repo: repo,
	}
}
func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment) error {
	if comment.Content == "" {
		return apperrors.ErrInvalidInput
	}
	return s.repo.CreateComment(ctx, comment)
}
func (s *CommentService) GetCommentsByPost(ctx context.Context, postID int) ([]models.Comment, error) {
	comments, err := s.repo.GetCommentsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	for i := range comments {
		if comments[i].IsDeleted() {
			s.maskDeletedComment(&comments[i])
		}
	}
	return comments, nil
}

// maskDeletedComment replaces the content and author of a soft-deleted
// comment with placeholders, so the deleted comment never exposes its author
// or content. It mirrors how deleted posts keep only a "[deleted]" title.
func (s *CommentService) maskDeletedComment(comment *models.Comment) {
	comment.Content = "[deleted]"
	comment.AuthorName = ""
	comment.AuthorPicture = ""
	comment.UserID = 0
}

// GetCommentThreads returns the top-level comments of a post (newest first).
// Each comment carries its direct replies (oldest first), which carry their
// own replies, so every reply sits under the comment it answers.
func (s *CommentService) GetCommentThreads(ctx context.Context, postID int) ([]models.Comment, error) {
	comments, err := s.repo.GetCommentsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	exists := make(map[int]bool, len(comments))
	for _, c := range comments {
		exists[c.ID] = true
	}

	for i := range comments {
		if comments[i].IsDeleted() {
			s.maskDeletedComment(&comments[i])
		}
	}

	var top []models.Comment
	for _, c := range comments {
		if c.ParentID == nil || !exists[*c.ParentID] {
			top = append(top, c)
		}
	}

	// comments come newest first, so walk backwards to list replies oldest first
	children := make(map[int][]models.Comment)
	for i := len(comments) - 1; i >= 0; i-- {
		c := comments[i]
		if c.ParentID != nil && exists[*c.ParentID] {
			children[*c.ParentID] = append(children[*c.ParentID], c)
		}
	}

	var attach func(c *models.Comment)
	attach = func(c *models.Comment) {
		c.Replies = children[c.ID]
		for i := range c.Replies {
			attach(&c.Replies[i])
		}
	}
	for i := range top {
		attach(&top[i])
	}
	return top, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID int, userID int) (*models.Comment, error) {
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.DeletedAt != nil {
		return nil, apperrors.ErrNotFound
	}

	if comment.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if err := s.repo.DeleteComment(ctx, commentID); err != nil {
		return nil, err
	}

	// Soft delete: the comment becomes a deleted comment, but it stays in the
	// thread so its replies and the other comments are kept, just like deleted
	// posts keep their comments.
	now := time.Now()
	comment.DeletedAt = &now
	s.maskDeletedComment(comment)

	return comment, nil
}

func (s *CommentService) GetCommentCount(ctx context.Context, postID int) (int, error) {
	return s.repo.GetCommentCount(ctx, postID)
}
