package service

import (
	"context"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
	"strings"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) FindUserById(ctx context.Context, id int) (*models.User, error) {
	return s.repo.FindUserById(ctx, id)
}

func (s *UserService) UpdateProfilePicture(ctx context.Context, userID int, picturePath string) error {
	return s.repo.UpdateProfilePicture(ctx, userID, picturePath)
}

func (s *UserService) UpdateUserName(ctx context.Context, userID int, username string) error {
	username = strings.TrimSpace(username)
	if err := validateUserName(username); err != nil {
		return err
	}
	return s.repo.UpdateUserName(ctx, userID, username)
}
