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
func (s *PostService) CreatePost(ctx context.Context, post *models.Post, categoryID int) error {
	if post.Content == "" {
		return apperrors.ErrInvalidInput
	}
	if categoryID < 1{
		return apperrors.ErrInvalidInput
	}
	return s.repo.CreatePost(ctx, post, categoryID)
}

func (s *PostService) GetAllPosts( ctx context.Context)([]models.Post, error){
	return s.repo.GetAllPosts(ctx)
}

func(s *PostService) GetPostByUser(ctx context.Context, userId int)([]models.Post, error){
	return s.repo.GetPostByUser(ctx, userId)
}