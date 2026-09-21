package repository

import (
	"context"
	"database/sql"
	"errors"
	"uuid"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?,?,?)
	`, session.ID.String(), session.UserID, session.ExpiresAt)

	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	session := &models.Session{}

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, expires_at 
		FROM session 
		WHERE id ?
	`, id.String()).Scan(
		&session.ID,
		&session.ExpiresAt,
		&session.UserID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE from sessions
		WHERE id ?
	`, id.String(),
	)

	return err
}
