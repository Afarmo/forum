package handlers

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type PostHandler struct {
	service *service.PostService
	tmpl    *template.Template
}

func NewPostHandler(service *service.PostService, tmpl *template.Template) *PostHandler {
	return &PostHandler{
		service: service,
		tmpl:    tmpl,
	}
}
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	category_id := r.FormValue("category_id")
	if category_id == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
		return
	}
	categoryID, err := strconv.Atoi(category_id)
	if err != nil {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}
	var post models.Post

	post.UserID = 1 // dummmy id - waiting for authentication to get the user id from the session
	post.Content = content
	file, header, err := r.FormFile("picture") // WIP
	if err == nil {

		defer file.Close()

		err = os.MkdirAll("internal/web/static/img/post", 0755)
		if err != nil {
			http.Error(w, "failed to create directory", http.StatusInternalServerError)
			return
		}
		destination, err := os.Create("internal/web/static/img/post/" + header.Filename)
		if err != nil {
			http.Error(w, "failed to create file", http.StatusInternalServerError)
			return
		}
		defer destination.Close()

		if _, err := io.Copy(destination, file); err != nil {
			http.Error(w, "falied to save uploaded file", http.StatusInternalServerError)
			return
		}

		post.PictureContent = header.Filename
	}

	err = h.service.CreatePost(ctx, &post, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	posts, err := h.service.GetAllPosts(ctx)
	if err != nil {
		log.Println("Get post error:", err)
		http.Error(w, "failed to get the posts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) GetPostByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	posts, err := h.service.GetPostByUser(ctx, userID)
	if err != nil {
		http.Error(w, "failed to get the user posts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	PostID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid Post ID", http.StatusBadRequest)
		return
	}
	var update models.UpdatePost

	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "invalid Json", http.StatusBadRequest)
		return
	}
	err = h.service.UpdatePost(ctx, &PostID, &update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post updated successfully"))
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}
	err = h.service.DeletePost(ctx, &postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post deleted successfully"))
}
