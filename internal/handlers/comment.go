package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type CommentHandler struct {
	service *service.CommentService
	tmpl    *template.Template
}

func NewCommentHandler(service *service.CommentService) *CommentHandler {
	return &CommentHandler{
		service: service,
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
	ctx := r.Context()
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

	var comment models.Comment
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
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

	err = h.service.CreateComment(ctx, &comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Land on the new comment (#comment-N) so the reload scrolls straight to it.
	http.Redirect(w, r, "/posts/"+strconv.Itoa(postID)+"#comment-"+strconv.Itoa(comment.ID), http.StatusSeeOther)

}

func (h *CommentHandler) GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	posts, err := h.service.GetCommentsByPost(ctx, postID)
	if err != nil {
		http.Error(w, "failed to get the user posts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}
