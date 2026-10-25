package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type CommentHandler struct {
	service *service.CommentService
	tmpl    *template.Template
}

func NewCommentHandler(service *service.CommentService, tmpl *template.Template) *CommentHandler {
	return &CommentHandler{
		service: service,
		tmpl:    tmpl,
	}
}

// CommentView is what the "comment" template renders: a comment, its replies,
// and whether the viewer may reply to it.
type CommentView struct {
	models.Comment
	Replies  []CommentView
	CanReply bool
}

func newCommentViews(comments []models.Comment, canReply bool) []CommentView {
	views := make([]CommentView, len(comments))
	for i, c := range comments {
		views[i] = CommentView{Comment: c, Replies: newCommentViews(c.Replies, canReply), CanReply: canReply}
	}
	return views
}

func countComments(views []CommentView) int {
	n := len(views)
	for _, v := range views {
		n += countComments(v.Replies)
	}
	return n
}
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())

	var comment models.Comment
	comment.UserID = user.ID
	comment.PostID = postID
	comment.Content = content

	parentID := r.FormValue("parent_id")
	if parentID != "" {
		id, err := strconv.Atoi(parentID)
		if err != nil {
			http.Error(w, "invalid parent id", http.StatusBadRequest)
			return
		}
		comment.ParentID = &id
	}

	if err = h.service.CreateComment(r.Context(), &comment); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidInput):
			http.Error(w, "invalid comment", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "post not found", http.StatusNotFound)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}

		return
	}

	// Land on the new comment (#comment-N) so the reload scrolls straight to it.
	http.Redirect(w, r, "/posts/"+strconv.Itoa(postID)+"#comment-"+strconv.Itoa(comment.ID), http.StatusSeeOther)
}

func (h *CommentHandler) GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	_, err = h.service.GetCommentsByPost(r.Context(), postID)
	if err != nil {
		apperrors.Log(err)
		http.Error(w, "failed to get comments", http.StatusInternalServerError)
		return
	}

}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	comment, err := h.service.DeleteComment(r.Context(), commentID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "comment not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrForbidden):
			http.Error(w, "you cannot delete this comment", http.StatusForbidden)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}
		return
	}
	postID := comment.PostID
	http.Redirect(w, r, "/posts/"+strconv.Itoa(postID)+"#comment-"+strconv.Itoa(commentID), http.StatusSeeOther)
}
