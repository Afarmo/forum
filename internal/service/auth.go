package service

import (
	"context"
	"net/mail"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/auth"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

const (
	minUsernameLength = 3
	maxUsernameLength = 20
	minPasswordLength = 8
	maxPasswordLength = 64
)

type AuthService struct {
	repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) error {
	if err := validateRegistration(username, email, password); err != nil {
		return err
	}

	hash, salt, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		UserName:     username,
		Email:        email,
		PasswordHash: hash,
		PasswordSalt: salt,
	}

	return s.repo.CreateUser(ctx, &user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) error {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return apperrors.ErrInvalidCredentials
	}

	if !auth.VerifyPassword(password, user.PasswordHash, user.PasswordSalt) {
		return apperrors.ErrInvalidCredentials
	}

	return nil
}

func validateRegistration(username, email, password string) error {
	if username == "" ||
		len(username) < minUsernameLength ||
		len(username) > maxUsernameLength {
		return apperrors.ErrInvalidUsername
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return apperrors.ErrInvalidEmail
	}

	if password == "" ||
		len(password) < minPasswordLength ||
		len(password) > maxPasswordLength {
		return apperrors.ErrInvalidPassword
	}

	return nil
}
