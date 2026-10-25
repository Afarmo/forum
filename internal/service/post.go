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

func (s *PostService) GetAllPosts(ctx context.Context, categoryID, currentUserID int) ([]models.Post, error) {
	return s.repo.GetAllPosts(ctx, categoryID, currentUserID)
}

func (s *PostService) GetPostByUser(ctx context.Context, userID, currentUserID int) ([]models.Post, error) {
	return s.repo.GetPostByUser(ctx, userID, currentUserID)
}

func (s *PostService) UpdatePost(ctx context.Context, PostID *int, update *models.UpdatePost) error {
	if PostID == nil || *PostID <= 0 {
		return apperrors.ErrNotFound
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

func (s *PostService) DeletePost(ctx context.Context, PostID *int) error {
	if PostID == nil || *PostID <= 0 {
		return apperrors.ErrInvalidInput
	}

	return s.repo.DeletePost(ctx, PostID)
}

func (s *PostService) GetPostById(ctx context.Context, postId, currentUserID int) (*models.Post, error) {
	if postId <= 0 {
		return nil, apperrors.ErrInvalidInput
	}

	return s.repo.GetPostById(ctx, postId, currentUserID)
}

func (s *PostService) SearchPosts(ctx context.Context, search string, currentUserID int) ([]models.Post, error) {
	if search == "" {
		return nil, apperrors.ErrNotFound
	}
	return s.repo.SearchPosts(ctx, search, currentUserID)
}
