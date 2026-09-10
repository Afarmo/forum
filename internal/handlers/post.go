package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
	}
	category_id := r.FormValue("category_id")
	if category_id == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
	}
	file, header, err := r.FormFile("picture") // WIP
	if err != nil {
		http.Error(w, "picture is required", http.StatusBadRequest)
	}
	defer file.Close()

	var post models.Post

	// if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
	// 	http.Error(w, "invalid JSON", http.StatusBadRequest)
	// 	return
	// }
	post.UserId = 1 // dummmy id - waiting for authentication to get the user id from the session
	post.Content = content
	post.PictureContent = header.Filename
	categoryID, err := strconv.Atoi(category_id)
	if err := h.service.CreatePost(ctx, &post, categoryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // TODO
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}
