package repository

import (
	"FORUM/internal/errorMsg"

	"github.com/Afarmo/forum/internal/models"

	"context"
	"database/sql"
	"strings"
	"time"
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
		return errorMsg.ErrTransactionStart
	}
	defer tx.Rollback() // if there is error revert changes to the db back to before the changes
	query := `INSERT INTO users(username, email, user_password) VALUES(?,?,?)`
	now := time.Now()
	result, err := tx.ExecContext(ctx, query, user.UserName, user.Email, user.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errorMsg.ErrDuplicateKey
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.Id = int(id)
	user.CreatedAt = now
	return tx.Commit()
}
