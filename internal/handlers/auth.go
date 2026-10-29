package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"uuid"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/auth"
	"github.com/Afarmo/forum/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
	tmpl    *template.Template
}

func NewAuthHandler(service *service.AuthService, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{
		service: service,
		tmpl:    tmpl,
	}
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if err := h.service.Register(r.Context(), username, email, password); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidUsername):
			http.Redirect(w, r, "/?register_error=invalid_username", http.StatusSeeOther)
		case errors.Is(err, apperrors.ErrInvalidEmail):
			http.Redirect(w, r, "/?register_error=invalid_email", http.StatusSeeOther)
		case errors.Is(err, apperrors.ErrInvalidPassword):
			http.Redirect(w, r, "/?register_error=invalid_password", http.StatusSeeOther)
		case errors.Is(err, apperrors.ErrDuplicateUsername):
			http.Redirect(w, r, "/?register_error=duplicate_username", http.StatusSeeOther)
		case errors.Is(err, apperrors.ErrDuplicateEmail):
			http.Redirect(w, r, "/?register_error=duplicate_email", http.StatusSeeOther)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	session, err := h.service.Login(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidCredentials):
			http.Redirect(w, r, "/?login_error=1", http.StatusSeeOther)

		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID.String(),
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/?login_success=1", http.StatusSeeOther)
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionID, err := uuid.Parse(cookie.Value)
		if err == nil {
			if err := h.service.Logout(r.Context(), sessionID); err != nil {
				apperrors.Log(err)
				auth.ClearSessionCookie(w)
				apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
				return
			}
		}
	}

	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
