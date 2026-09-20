package handlers

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/Afarmo/forum/internal/apperrors"
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
		apperrors.Log(err)
		switch {
		case errors.Is(err, apperrors.ErrInvalidUsername),
			errors.Is(err, apperrors.ErrInvalidEmail),
			errors.Is(err, apperrors.ErrInvalidPassword),
			errors.Is(err, apperrors.ErrDuplicateKey):
			http.Error(w, err.Error(), http.StatusBadRequest)

		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		}
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	session, err := h.service.Login(r.Context(), email, password)
	if err != nil {
		apperrors.Log(err)
		switch {
		case errors.Is(err, apperrors.ErrInvalidCredentials):
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)

		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID.String(),
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
