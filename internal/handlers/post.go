package handlers

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Afarmo/forum/internal/middleware"
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
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
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
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	post.UserID = user.ID
	post.Content = content
	post.Title = title
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

	categoryID := 0
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		var convErr error
		categoryID, convErr = strconv.Atoi(categoryIDStr)
		if convErr != nil {
			http.Error(w, "invalid category_id", http.StatusBadRequest)
			return
		}
	}

	posts, err := h.service.GetAllPosts(ctx, categoryID)
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

func (h *PostHandler) SearchPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	var posts []models.Post
	var err error
	if search == "" {
		posts, err = h.service.GetAllPosts(ctx, 0)
	} else {
		posts, err = h.service.SearchPosts(ctx, search)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}
