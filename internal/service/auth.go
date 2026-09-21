package service

import (
	"context"
	"net/mail"
	"time"
	"uuid"

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
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
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

	return s.userRepo.CreateUser(ctx, &user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.Session, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if !auth.VerifyPassword(password, user.PasswordHash, user.PasswordSalt) {
		return nil, apperrors.ErrInvalidCredentials
	}

	session := &models.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserID:    user.ID,
	}

	if err = s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
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
