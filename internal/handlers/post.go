package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type PostHandler struct {
	service *service.PostService
}

func NewPostHandler(service *service.PostService) *PostHandler {
	return &PostHandler{
		service: service,
	}
}
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var post models.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest) // TODO
		return
	}
	post.UserId = 1 // dummmy id - waiting for authentication to get the user id from the session
	if err := h.service.NewPostService(ctx, &post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // TODO
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}
