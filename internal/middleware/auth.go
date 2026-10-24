package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
	"uuid"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/auth"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/repository"
)

type AuthMiddleware struct {
	sessionRepo *repository.SessionRepository
	userRepo    *repository.UserRepository
}

func NewAuthMiddleware(
	sessionRepo *repository.SessionRepository,
	userRepo *repository.UserRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			auth.ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		session, err := m.sessionRepo.FindByID(r.Context(), sessionID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				auth.ClearSessionCookie(w)
			} else {
				apperrors.Log(err)
			}

			next.ServeHTTP(w, r)
			return
		}

		if !time.Now().Before(session.ExpiresAt) {
			if err := m.sessionRepo.DeleteSession(r.Context(), sessionID); err != nil {
				apperrors.Log(err)
			}

			auth.ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.userRepo.FindUserById(r.Context(), session.UserID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				auth.ClearSessionCookie(w)
				if err = m.sessionRepo.DeleteSession(r.Context(), sessionID); err != nil {
					apperrors.Log(err)
				}
			} else {
				apperrors.Log(err)
			}
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			log.Println("Unauthenticated: redirecting to /")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type userContextKey struct{}

func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(userContextKey{}).(*models.User)
	return user
}
