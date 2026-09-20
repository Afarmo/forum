package service

import (
	"context"
	"strings"

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
	if post.Title == "" {
		return apperrors.ErrInvalidInput
	}
	if post.Content == "" {
		return apperrors.ErrInvalidInput
	}
	if categoryID < 1 {
		return apperrors.ErrInvalidInput
	}
	return s.repo.CreatePost(ctx, post, categoryID)
}

func (s *PostService) GetAllPosts(ctx context.Context, categoryID int) ([]models.Post, error) {
	return s.repo.GetAllPosts(ctx, categoryID)
}

func (s *PostService) GetPostByUser(ctx context.Context, userId int) ([]models.Post, error) {
	return s.repo.GetPostByUser(ctx, userId)
}

func (s *PostService) UpdatePost(ctx context.Context, PostID *int, update *models.UpdatePost) error {
	if *PostID <= 0 || PostID == nil{
		return apperrors.ErrInvalidInput
	}
	if update.Title == nil && update.Content == nil && update.PictureContent == nil {
		return apperrors.ErrInvalidInput
	}
	if update.Content != nil && strings.TrimSpace(*update.Content) == "" {
		return apperrors.ErrInvalidInput
	}
	if update.Title != nil && strings.TrimSpace(*update.Title) == "" {
		return apperrors.ErrInvalidInput
	}
	return s.repo.UpdatePost(ctx, *PostID, update)
}

func (s *PostService) DeletePost(ctx context.Context, PostID *int) error{
	if  *PostID <= 0 || PostID == nil{
		return apperrors.ErrInvalidInput
	}
	return s.repo.DeletePost(ctx, PostID)
}