package repository

import (
	"context"
	"database/sql"
	"time"
	"apperrors"

	"github.com/Afarmo/forum/internal/models"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *models.Comment) error {

	query := `INSERT INTO comments(user_id, post_id, parent_id, content) VALUES(?,?,?,?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, comment.UserID, comment.PostID, comment.ParentID, comment.Content)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	comment.ID = int(id)
	comment.CreatedAt = now
	comment.UpdatedAt = now
	return nil
}

func (r *CommentRepository) GetCommentsByPost(ctx context.Context, postID int) ([]models.Comment, error) {
	query := `SELECT id, user_id, post_id, parent_id, content, created_at, updated_at FROM comments WHERE post_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, commentID int) (*models.Comment, error) {
	query := `SELECT id, user_id, post_id, parent_id, content, created_at, updated_at FROM comments WHERE id = ?`

	var comment models.Comment
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
