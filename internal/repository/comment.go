package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Afarmo/forum/internal/apperrors"
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

	query := `INSERT INTO comments(user_id, post_id, parent_id, content) SELECT ?, ?, ?, ? 
		WHERE EXISTS (SELECT 1 FROM posts WHERE id = ? AND deleted_at IS NULL)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, comment.UserID, comment.PostID, comment.ParentID, comment.Content, comment.PostID)
	if err != nil {
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return apperrors.ErrNotFound
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
	query := `SELECT c.id, c.user_id, c.post_id, c.parent_id, c.content, c.created_at, c.updated_at, c.deleted_at,
			COALESCE(u.username, ''), COALESCE(u.profile_picture, '')
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC, c.id DESC`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
			&comment.AuthorName, &comment.AuthorPicture)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, commentID int) (*models.Comment, error) {
	query := `SELECT id, user_id, post_id, parent_id, content, created_at, updated_at, deleted_at FROM comments WHERE id = ?`

	var comment models.Comment
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
func (r *CommentRepository) DeleteComment(ctx context.Context, commentID int) error {
	query := `
        UPDATE comments
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE id = ? AND deleted_at IS NULL
    `

	result, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	// Only the comment itself is soft deleted; replies to it are kept intact.

	_, err = r.db.ExecContext(ctx, `DELETE FROM comment_reactions WHERE comment_id = ?`, commentID)
	if err != nil {
		return err
	}

	return nil
}
