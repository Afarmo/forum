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
	tx, txErr := r.db.BeginTx(ctx, nil)
	if txErr != nil {
		return apperrors.ErrTransactionStart
	}
	defer tx.Rollback() // if there is error revert changes to the db back to before the changes
	query := `INSERT INTO users(username, email, user_password, profile_picture) VALUES(?,?,?,?)`
	now := time.Now()
	result, err := tx.ExecContext(ctx, query, user.UserName, user.Email, user.Password, user.ProfilePicture)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return apperrors.ErrDuplicateKey
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)
	user.CreatedAt = now
	return tx.Commit()
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, user_password, profile_picture, created_at FROM USERS WHERE email = ?`
	user := &models.User{}

	var profilePicture sql.NullString
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Password,
		&profilePicture,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindUserById(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, email, user_password, profile_picture, created_at FROM USERS WHERE id = ?`
	user := &models.User{}

	var profilePicture sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Password,
		&profilePicture,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
func (r *UserRepository) UpdateProfilePicture(ctx context.Context, userID int, picturePath string) error {
	query := `UPDATE users SET profile_picture = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, picturePath, userID)
	return err
}
