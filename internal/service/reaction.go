package service

import (
	"context"
	"errors"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type ReactionService struct {
	reactionRepo *repository.ReactionRepository
}

func NewReactionService(reactionRepo *repository.ReactionRepository) *ReactionService {
	return &ReactionService{
		reactionRepo: reactionRepo,
	}
}

func (s *ReactionService) ReactToPost(ctx context.Context, postID, userID, reaction int) error {
	if reaction != 1 && reaction != -1 {
		return apperrors.ErrInvalidInput
	}

	current, err := s.reactionRepo.GetPostReaction(ctx, postID, userID)

	if errors.Is(err, apperrors.ErrNotFound) {
		return s.reactionRepo.CreatePostReaction(ctx, &models.PostReaction{
			PostID:   postID,
			UserID:   userID,
			Reaction: reaction,
		})
	}

	if err != nil {
		return err
	}

	if current.Reaction == reaction {
		return s.reactionRepo.DeletePostReaction(ctx, postID, userID)
	}

	current.Reaction = reaction
	return s.reactionRepo.UpdatePostReaction(ctx, current)
}

func (s *ReactionService) GetPostReactionCount(ctx context.Context, postID int) (int, int, error) {
	return s.reactionRepo.GetPostReactionCount(ctx, postID)
}

func (s *ReactionService) ReactToComment(ctx context.Context, commentID, userID, reaction int) error {
	if reaction != 1 && reaction != -1 {
		return apperrors.ErrInvalidInput
	}

	current, err := s.reactionRepo.GetCommentReaction(ctx, commentID, userID)

	if errors.Is(err, apperrors.ErrNotFound) {
		return s.reactionRepo.CreateCommentReaction(
			ctx,
			&models.CommentReaction{
				CommentID: commentID,
				UserID:    userID,
				Reaction:  reaction,
			},
		)
	}

	if err != nil {
		return err
	}

	if current.Reaction == reaction {
		return s.reactionRepo.DeleteCommentReaction(ctx, commentID, userID)
	}

	current.Reaction = reaction
	return s.reactionRepo.UpdateCommentReaction(ctx, current)
}

func (s *ReactionService) GetCommentReactionCount(ctx context.Context, commentID int) (int, int, error) {
	return s.reactionRepo.GetCommentReactionCount(ctx, commentID)
}

func (s *ReactionService) GetCommentReaction(ctx context.Context, postID, userID int) (*models.CommentReaction, error) {
	return s.reactionRepo.GetCommentReaction(ctx, postID, userID)
}
