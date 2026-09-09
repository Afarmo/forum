package service

import (
	"context"

	"github.com/Afarmo/forum/internal/errorMsg"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if user.UserName == "" || user.Email == "" {
		return errorMsg.ErrInvalidInput
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, errorMsg.ErrInvalidInput
	}
	return s.repo.FindUserByEmail(ctx, email)
}
func (s *UserService) FindUserById(ctx context.Context, id int) (*models.User, error) {
	if id < 1 {
		return nil, errorMsg.ErrInvalidInput
	}
	return s.repo.FindUserById(ctx, id)
}
