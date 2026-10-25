package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/service"
)

type ReactionHandler struct {
	reactionService *service.ReactionService
	tmpl            *template.Template
}

func NewReactionHandler(reactionService *service.ReactionService, tmpl *template.Template) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
		tmpl:            tmpl,
	}
}

func (h *ReactionHandler) ReactToPost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || postID <= 0 {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	reaction, err := strconv.Atoi(r.FormValue("reaction"))
	if err != nil {
		http.Error(w, "invalid reaction", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())

	if err := h.reactionService.ReactToPost(
		r.Context(),
		postID,
		user.ID,
		reaction,
	); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidInput):
			http.Error(w, "invalid reaction", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "not found", http.StatusNotFound)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}

		return
	}

	http.Redirect(w, r, "/posts/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func (h *ReactionHandler) ReactToComment(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.PathValue("postID"))
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	reaction, err := strconv.Atoi(r.FormValue("reaction"))
	if err != nil {
		http.Error(w, "invalid reaction", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(r.PathValue("commentID"))
	if err != nil {
		http.Error(w, "invalid comment ID", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())

	if err := h.reactionService.ReactToComment(
		r.Context(),
		commentID,
		user.ID,
		reaction,
	); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidInput):
			http.Error(w, "invalid reaction", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "not found", http.StatusNotFound)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}

		return
	}

	http.Redirect(w, r, "/posts/"+strconv.Itoa(postID), http.StatusSeeOther)
}
