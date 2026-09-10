package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{
		db: db,
	}
}

func (r *PostRepository) CreatePost(ctx context.Context, post *models.Post) error {
	tx, txErr := r.db.BeginTx(ctx, nil)
	if txErr != nil {
		return apperrors.ErrTransactionStart
	}
	defer tx.Rollback() // if there is error revert changes to the db back to before the changes
	query := `INSERT INTO posts(user_id, Content, picture_content) VALUES(?,?,?)`
	now := time.Now()
	result, err := tx.ExecContext(ctx, query, post.UserId, post.Content, post.PictureContent)
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
	post.PostId = int(id)
	post.CreatedAt = now
	return tx.Commit()
}
