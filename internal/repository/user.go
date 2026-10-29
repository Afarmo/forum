package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users(
			username,
			email,
			password_hash,
			password_salt,
			profile_picture
		) VALUES(?,?,?,?,?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		query,
		user.UserName,
		user.Email,
		user.PasswordHash,
		user.PasswordSalt,
		user.ProfilePicture,
	)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "UNIQUE constraint failed: USERS.email"):
			return apperrors.ErrDuplicateEmail
		case strings.Contains(err.Error(), "UNIQUE constraint failed: USERS.username"):
			return apperrors.ErrDuplicateUsername
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	user.CreatedAt = now

	return nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, password_salt, profile_picture, created_at FROM USERS WHERE email = ?`
	user := &models.User{}

	var profilePicture sql.NullString
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.PasswordHash,
		&user.PasswordSalt,
		&profilePicture,
		&user.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	user.ProfilePicture = profilePicture.String

	return user, nil
}

func (r *UserRepository) FindUserById(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, password_salt, profile_picture, created_at FROM USERS WHERE id = ?`
	user := &models.User{}

	var profilePicture sql.NullString
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.PasswordHash,
		&user.PasswordSalt,
		&profilePicture,
		&user.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	user.ProfilePicture = profilePicture.String

	return user, nil
}

func (r *UserRepository) UpdateProfilePicture(ctx context.Context, userID int, picturePath string) error {
	query := `UPDATE users SET profile_picture = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, picturePath, userID)
	return err
}
func (r *UserRepository) UpdateUserName(ctx context.Context, userID int, username string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	query := `UPDATE users SET username = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, username, userID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return apperrors.ErrDuplicateKey
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	_, err = tx.ExecContext(ctx, `UPDATE POSTS_SEARCH SET username = ? WHERE rowid IN (SELECT id FROM posts WHERE user_id = ?)`, username, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
