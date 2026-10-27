package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type PostHandler struct {
	service        *service.PostService
	commentService *service.CommentService
	tmpl           *template.Template
}

func NewPostHandler(service *service.PostService, tmpl *template.Template, commentService *service.CommentService) *PostHandler {
	return &PostHandler{
		service:        service,
		commentService: commentService,
		tmpl:           tmpl,
	}
}

type PostPageData struct {
	Title        string
	User         *models.User
	Categories   []models.Category
	Post         *models.Post
	Comments     []CommentView
	CommentCount int
	SearchQuery  string
}

func (h *PostHandler) CreatePostPage(w http.ResponseWriter, r *http.Request) {

	num, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	post, err := h.service.GetPostById(ctx, num)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			apperrors.RenderError(w, h.tmpl, http.StatusNotFound)
			return
		}

		log.Println("get post error:", err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	comments, err := h.commentService.GetCommentThreads(ctx, num)
	if err != nil {
		log.Println("get comments error:", err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	user := middleware.UserFromContext(ctx)
	viewerID := 0

	if user != nil {
		viewerID = user.ID
	}

	views := newCommentViews(comments, user != nil, viewerID)

	data := &PostPageData{
		Title:        post.Title,
		User:         user,
		Post:         post,
		Comments:     views,
		CommentCount: countComments(views),
	}

	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, "post_page.html", data); err != nil {
		log.Println("error rendering template:", err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Println("error writing response:", err)
	}

}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
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

	post.Title = title
	post.Content = content
	post.UserID = user.ID

	file, header, err := r.FormFile("picture")
	if err == nil {
		defer file.Close()

		err = os.MkdirAll("internal/web/static/img/post", 0755)
		if err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}
		destination, err := os.Create("internal/web/static/img/post/" + header.Filename)
		if err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}
		defer destination.Close()

		if _, err := io.Copy(destination, file); err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}

		post.PictureContent = header.Filename
	} else if imageURL := strings.TrimSpace(r.FormValue("image_url")); imageURL != "" {
		u, err := url.Parse(imageURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			http.Error(w, "image URL must be a full http:// or https:// link", http.StatusBadRequest)
			return
		}
		post.PictureContent = imageURL
	}

	err = h.service.CreatePost(r.Context(), &post, categoryID)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidInput) {
			http.Error(w, "invalid post data", http.StatusBadRequest)
			return
		}
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
	}
	// Back to the feed, scrolled to the new post's card (#post-N).
	http.Redirect(w, r, "/#post-"+strconv.Itoa(post.PostID), http.StatusSeeOther)
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	var categoryID int
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		var err error
		categoryID, err = strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "invalid category id", http.StatusBadRequest)
			return
		}
	}

	posts, err := h.service.GetAllPosts(r.Context(), categoryID)
	if err != nil {
		apperrors.Log(err)
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
		apperrors.Log(err)
		http.Error(w, "failed to get the user posts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	PostID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid Post ID", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var update models.UpdatePost

	update.Title = &title
	update.Content = &content

	err = h.service.UpdatePost(r.Context(), &PostID, user.ID, &update)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "post not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrInvalidInput):
			http.Error(w, "invalid post data", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrForbidden):
			http.Error(w, "you cannot edit this post", http.StatusForbidden)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%d", user.ID), http.StatusSeeOther)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.PathValue("id"))
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	if err = h.service.DeletePost(r.Context(), &postID, user.ID); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidInput):
			http.Error(w, "invalid post ID", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "post not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrForbidden):
			http.Error(w, "you cannot delete this post", http.StatusForbidden)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/users/"+strconv.Itoa(user.ID), http.StatusSeeOther)
}

func (h *PostHandler) SearchPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := strings.TrimSpace(r.URL.Query().Get("query"))
	var posts []models.Post
	var err error
	if search == "" {
		posts, err = h.service.GetAllPosts(ctx, 0)
	} else {
		posts, err = h.service.SearchPosts(ctx, search)
	}
	if err != nil {
		apperrors.Log(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []models.Post{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}
