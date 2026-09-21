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
