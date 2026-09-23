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
		VALUES(?,?,?)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		reaction.PostID,
		reaction.UserID,
		reaction.Reaction,
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
