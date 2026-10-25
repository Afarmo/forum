package handlers

import (
	"encoding/json"
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

	if r.Header.Get("Accept") == "application/json" {
		likes, dislikes, err := h.reactionService.GetPostReactionCount(r.Context(), postID)
		if err != nil {
			apperrors.Log(err)
			http.Error(w, "failed to load updated reaction counts", http.StatusInternalServerError)
			return
		}

		userReaction := 0
		current, err := h.reactionService.GetPostReaction(r.Context(), postID, user.ID)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			apperrors.Log(err)
			http.Error(w, "failed to load updated reaction state", http.StatusInternalServerError)
			return
		}
		if current != nil {
			userReaction = current.Reaction
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			LikeCount    int `json:"like_count"`
			DislikeCount int `json:"dislike_count"`
			UserReaction int `json:"user_reaction"`
		}{
			LikeCount:    likes,
			DislikeCount: dislikes,
			UserReaction: userReaction,
		}); err != nil {
			apperrors.Log(err)
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
