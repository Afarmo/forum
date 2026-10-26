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

func (s *PostService) GetPostByUser(ctx context.Context, userID int) ([]models.Post, error) {
	return s.repo.GetPostByUser(ctx, userID)
}

func (s *PostService) UpdatePost(ctx context.Context, PostID *int, userID int, update *models.UpdatePost) error {
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
	post, err := s.repo.GetPostById(ctx, *PostID)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return apperrors.ErrNotFound
	}
	if post.UserID != userID {
		return apperrors.ErrForbidden
	}

	return s.repo.UpdatePost(ctx, *PostID, userID, update)
}

func (s *PostService) DeletePost(ctx context.Context, PostID *int, userID int) error {
	if PostID == nil || *PostID <= 0 {
		return apperrors.ErrInvalidInput
	}

	post, err := s.repo.GetPostById(ctx, *PostID)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return apperrors.ErrForbidden
	}
	return s.repo.DeletePost(ctx, PostID, userID)
}

func (s *PostService) GetPostById(ctx context.Context, postId, currentUserID int) (*models.Post, error) {
	if postId <= 0 {
		return nil, apperrors.ErrInvalidInput
	}

	post, err := s.repo.GetPostById(ctx, postId)
	if err != nil {
		return nil, err
	}

	// Deleted posts keep their comments, but the content never leaves the backend.
	if post.IsDeleted() {
		post.Title = "[deleted]"
		post.Content = ""
		post.PictureContent = ""
		post.AuthorName = ""
		post.AuthorPicture = ""
		post.UserID = 0
	}
	return post, nil
}

func (s *PostService) SearchPosts(ctx context.Context, search string) ([]models.Post, error) {
	if search == "" {
		return nil, apperrors.ErrNotFound
	}
	return s.repo.SearchPosts(ctx, search)
}
