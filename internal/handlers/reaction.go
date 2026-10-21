package handlers

import (
	"net/http"
	"strconv"

	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/service"
)

type ReactionHandler struct {
	reactionService *service.ReactionService
}

func NewReactionHandler(reactionService *service.ReactionService) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
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
		return
	}

	http.Redirect(w, r, "/posts/"+strconv.Itoa(postID), http.StatusSeeOther)
}
