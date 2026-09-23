package service

import (
	"context"

	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetAllCategories(ctx)
}
func (s *CategoryService) GetCategoryByUser(ctx context.Context, userId int) ([]models.Category, error) {
	return s.repo.GetCategoryByUser(ctx, userId)
}
