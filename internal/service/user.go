package service

import (
	"FORUM/internal/errorMsg"
	"context"

	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserRepository(repo *repository.UserRepository) *UserService {
	return &UserRepository{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if user.UserName == "" || user.Email == "" {
		return errorMsg.ErrInvalidInput
	}
	return s.repo.CreateUser(ctx, user)
}
