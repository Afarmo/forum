package service

import (
	"context"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type PostService struct {
	repo *repository.PostRepository
}

func NewPostService(repo *repository.PostRepository) *PostService {
	return &PostService{
		repo: repo,
	}
}
func (s *PostService) NewPostService(ctx context.Context, post *models.Post) error {
	if post.Content == "" {
		return apperrors.ErrInvalidInput
	}
	return s.repo.CreatePost(ctx, post)
}
