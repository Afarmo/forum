package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
)

type ReactionRepository struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) *ReactionRepository {
	return &ReactionRepository{
		db: db,
	}
}

func (r *ReactionRepository) GetPostReaction(ctx context.Context, postID, userID int) (*models.PostReaction, error) {
	query := `
		SELECT post_id, user_id, reaction 
		FROM post_reactions 
		WHERE post_id = ? AND user_id = ?
	`
	reaction := &models.PostReaction{}
	if err := r.db.QueryRowContext(
		ctx,
		query,
		postID,
		userID,
	).Scan(
		&reaction.PostID,
		&reaction.UserID,
		&reaction.Reaction,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	return reaction, nil
}

func (r *ReactionRepository) CreatePostReaction(ctx context.Context, reaction *models.PostReaction) error {
	query := `
		INSERT INTO post_reactions(post_id, user_id, reaction) 
		SELECT ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM posts WHERE id = ? AND deleted_at IS NULL)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		reaction.PostID,
		reaction.UserID,
		reaction.Reaction,
		reaction.PostID,
	)

	return err
}

func (r *ReactionRepository) UpdatePostReaction(ctx context.Context, reaction *models.PostReaction) error {
	query := `
		UPDATE post_reactions 
		SET reaction = ? 
		WHERE post_id = ? AND user_id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reaction.Reaction,
		reaction.PostID,
		reaction.UserID,
	)

	return err
}

func (r *ReactionRepository) DeletePostReaction(ctx context.Context, postID, userID int) error {
	query := `
		DELETE FROM post_reactions 
		WHERE post_id = ? AND user_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, postID, userID)

	return err
}

func (r *ReactionRepository) GetPostReactionCount(ctx context.Context, postID int) (int, int, error) {
	query := `
		SELECT
			COUNT(CASE WHEN reaction = 1 THEN 1 END),
			COUNT(CASE WHEN reaction = -1 THEN 1 END)
		FROM post_reactions
		WHERE post_id = ?
	`

	var likes int
	var dislikes int

	err := r.db.QueryRowContext(ctx, query, postID).Scan(
		&likes,
		&dislikes,
	)

	return likes, dislikes, err
}

// ----- Comment Reactions -----

func (r *ReactionRepository) GetCommentReaction(ctx context.Context, commentID, userID int) (*models.CommentReaction, error) {
	query := `
		SELECT comment_id, user_id, reaction 
		FROM comment_reactions 
		WHERE comment_id = ? AND user_id = ?
	`
	reaction := &models.CommentReaction{}
	if err := r.db.QueryRowContext(
		ctx,
		query,
		commentID,
		userID,
	).Scan(
		&reaction.CommentID,
		&reaction.UserID,
		&reaction.Reaction,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	return reaction, nil
}

func (r *ReactionRepository) CreateCommentReaction(ctx context.Context, reaction *models.CommentReaction) error {
	query := `
		INSERT INTO comment_reactions(comment_id, user_id, reaction) 
		SELECT ?, ?, ?
		WHERE EXISTS (
			SELECT 1 FROM comments c
			JOIN posts p ON p.id = c.post_id
			WHERE c.id = ? AND c.deleted_at IS NULL AND p.deleted_at IS NULL
		)

	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		reaction.CommentID,
		reaction.UserID,
		reaction.Reaction,
		reaction.CommentID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *ReactionRepository) UpdateCommentReaction(ctx context.Context, reaction *models.CommentReaction) error {
	query := `
		UPDATE comment_reactions 
		SET reaction = ? 
		WHERE comment_id = ? AND user_id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reaction.Reaction,
		reaction.CommentID,
		reaction.UserID,
	)

	return err
}

func (r *ReactionRepository) DeleteCommentReaction(ctx context.Context, commentID, userID int) error {
	query := `
		DELETE FROM comment_reactions 
		WHERE comment_id = ? AND user_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, commentID, userID)

	return err
}

func (r *ReactionRepository) GetCommentReactionCount(ctx context.Context, commentID int) (int, int, error) {
	query := `
		SELECT
			COUNT(CASE WHEN reaction = 1 THEN 1 END),
			COUNT(CASE WHEN reaction = -1 THEN 1 END)
		FROM comment_reactions
		WHERE comment_id = ?
	`

	var likes int
	var dislikes int

	err := r.db.QueryRowContext(ctx, query, commentID).Scan(
		&likes,
		&dislikes,
	)

	return likes, dislikes, err
}
