package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
	"github.com/Afarmo/forum/internal/upload"
)

type PostHandler struct {
	postService     *service.PostService
	commentService  *service.CommentService
	reactionService *service.ReactionService
	tmpl            *template.Template
}

func NewPostHandler(postService *service.PostService,
	commentService *service.CommentService,
	reactionService *service.ReactionService,
	tmpl *template.Template,
) *PostHandler {
	return &PostHandler{
		postService:     postService,
		commentService:  commentService,
		reactionService: reactionService,
		tmpl:            tmpl,
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

	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	post, err := h.postService.GetPostById(ctx, postID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			apperrors.RenderError(w, h.tmpl, http.StatusNotFound)
			return
		}

		log.Println("get post error:", err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	comments, err := h.commentService.GetCommentThreads(ctx, postID)
	if err != nil {
		log.Println("get comments error:", err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}
	var userID int
	user := middleware.UserFromContext(ctx)
	if user != nil {
		userID = user.ID
	}

	views := newCommentViews(comments, user != nil, userID)

	posts := []models.Post{*post}

	if err := h.reactionService.PopulatePostStats(ctx, posts, userID); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	post = &posts[0]

	data := &PostPageData{
		Title:        post.Title,
		User:         user,
		Post:         post,
		Comments:     views,
		CommentCount: post.CommentCount,
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
	r.Body = http.MaxBytesReader(w, r.Body, upload.MaxRequestSize)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "request exceeds upload limit", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "invalid form data", http.StatusBadRequest)
		}
		return
	}

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

	categoryIDStr := r.FormValue("category_id")
	if categoryIDStr == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}

	var post models.Post
	user := middleware.UserFromContext(r.Context())

	post.Title = title
	post.Content = content
	post.UserID = user.ID

	postDir := filepath.Join("internal", "web", "static", "img", "post")

	file, header, err := r.FormFile("picture")
	if err == nil {
		defer file.Close()

		// Validate and save the uploaded image with a unique filename.
		filename, err := upload.SaveImage(file, header.Size, postDir)
		if err != nil {
			switch {
			case errors.Is(err, apperrors.ErrImageTooLarge):
				http.Error(w, "post image exceeds upload limit (5MiB)", http.StatusRequestEntityTooLarge)
			case errors.Is(err, apperrors.ErrInvalidImage):
				http.Error(w, "unsupported image format", http.StatusBadRequest)
			default:
				apperrors.Log(err)
				apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			}
			return
		}

		post.PictureContent = filename

	} else if errors.Is(err, http.ErrMissingFile) {
		// Allow an external image URL when no file was uploaded.
		if imageURL := strings.TrimSpace(r.FormValue("image_url")); imageURL != "" {
			u, err := url.Parse(imageURL)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				http.Error(w, "image URL must be a full http:// or https:// link", http.StatusBadRequest)
				return
			}

			post.PictureContent = imageURL
		}
	} else {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "request exceeds upload limit", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "failed to read uploaded image", http.StatusBadRequest)
		}
		return
	}

	if err := h.postService.CreatePost(r.Context(), &post, categoryID); err != nil {
		// Remove the uploaded image if the post couldn't be created.
		if file != nil {
			if removeErr := os.Remove(filepath.Join(postDir, post.PictureContent)); removeErr != nil {
				apperrors.Log(removeErr)
			}
		}

		if errors.Is(err, apperrors.ErrInvalidInput) {
			http.Error(w, "invalid post data", http.StatusBadRequest)
			return
		}

		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
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

	posts, err := h.postService.GetAllPosts(r.Context(), categoryID)
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
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	posts, err := h.postService.GetPostByUser(r.Context(), userID)
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

	err = h.postService.UpdatePost(r.Context(), &PostID, user.ID, &update)
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

	if err = h.postService.DeletePost(r.Context(), &postID, user.ID); err != nil {
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
		posts, err = h.postService.GetAllPosts(ctx, 0)
	} else {
		posts, err = h.postService.SearchPosts(ctx, search)
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
