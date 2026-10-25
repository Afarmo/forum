package service

import (
	"context"

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
	return s.repo.GetCommentsByPost(ctx, postID)
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

	return comment, nil
}
