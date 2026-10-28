package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"uuid"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type UserHandler struct {
	categoryService *service.CategoryService
	postService     *service.PostService
	userService     *service.UserService
	reactionService *service.ReactionService
	tmpl            *template.Template
}

func NewUserHandler(userService *service.UserService,
	categoryService *service.CategoryService,
	postService *service.PostService,
	reactionService *service.ReactionService,
	tmpl *template.Template,
) *UserHandler {
	return &UserHandler{
		userService:     userService,
		categoryService: categoryService,
		postService:     postService,
		reactionService: reactionService,
		tmpl:            tmpl,
	}
}

type UserPageData struct {
	Title                      string
	User                       *models.User
	ProfileUser                *models.User
	Categories                 []models.Category
	Posts                      []models.Post
	LikedPosts                 []models.Post
	SearchQuery                string
	ProfilePictureNotification string
}

func (h *UserHandler) FindUserById(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || userID < 1 {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	profileUser, err := h.userService.FindUserById(r.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			apperrors.RenderError(w, h.tmpl, http.StatusNotFound)
			return
		}

		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	profileUserPosts, err := h.postService.GetPostByUser(r.Context(), userID)
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	profileUserLikedPosts, err := h.postService.GetLikedPostsByUser(r.Context(), userID)
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	var loggedUserID int
	user := middleware.UserFromContext(r.Context())
	if user != nil {
		loggedUserID = user.ID
	}

	if err := h.reactionService.PopulatePostStats(r.Context(), profileUserPosts, loggedUserID); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	if err := h.reactionService.PopulatePostStats(r.Context(), profileUserLikedPosts, loggedUserID); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	data := &UserPageData{
		Title:       "User",
		User:        user,
		ProfileUser: profileUser,
		Categories:  categories,
		Posts:       profileUserPosts,
		LikedPosts:  profileUserLikedPosts,
	}
	if value := r.URL.Query().Get("profile_picture_status"); value == "success" {
		data.ProfilePictureNotification = "Successfully updated your profile picture."
	}

	var buf bytes.Buffer

	if err := h.tmpl.ExecuteTemplate(&buf, "user", data); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		apperrors.Log(err)
		return
	}
}

func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	const maxFileSize = 5 << 20
	const maxRequestSize = maxFileSize + 1<<20
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	file, header, err := r.FormFile("profile_picture")
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "request exceeds upload limit", http.StatusRequestEntityTooLarge)
			return
		}

		http.Error(w, "failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		http.Error(w, "profile picture exceeds upload limit (5MiB)", http.StatusRequestEntityTooLarge)
		return
	}

	// Check the file contents instead of trusting its extension
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		http.Error(w, "failed to read provided file", http.StatusBadRequest)
		return
	}

	// Identify the MIME type (<type>/<subtype>)
	contentType := http.DetectContentType(buf[:n])

	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpeg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		http.Error(w, "unsupported image format", http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	profileDir := filepath.Join("internal", "web", "static", "img", "profile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	filename := uuid.New().String() + ext
	picturePath := filepath.Join(profileDir, filename)

	destination, err := os.Create(picturePath)
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	var saved bool

	defer func() {
		if !saved {
			if err := os.Remove(picturePath); err != nil {
				apperrors.Log(err)
			}
		}
	}()

	defer destination.Close()

	if _, err = io.Copy(destination, file); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	if err := destination.Close(); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	pictureURL := "/static/img/profile/" + url.PathEscape(filename)

	if err := h.userService.UpdateProfilePicture(r.Context(), user.ID, pictureURL); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	saved = true

	http.Redirect(w, r, fmt.Sprintf("/users/%d?profile_picture_status=success", user.ID), http.StatusSeeOther)
}

func (h *UserHandler) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if err := h.userService.UpdateUserName(r.Context(), user.ID, r.FormValue("username")); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidUsername):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrDuplicateKey):
			http.Error(w, "username already taken", http.StatusConflict)
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "user not found", http.StatusNotFound)
		default:
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%d?profile_picture_status=success", user.ID), http.StatusSeeOther)
}
