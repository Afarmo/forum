package handlers

import (
	"bytes"
	"database/sql"
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
	tmpl            *template.Template
}

func NewUserHandler(userService *service.UserService, categoryService *service.CategoryService, postService *service.PostService, tmpl *template.Template) *UserHandler {
	return &UserHandler{
		userService:     userService,
		categoryService: categoryService,
		postService:     postService,
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

	viewer := middleware.UserFromContext(r.Context())
	currentUserID := middleware.UserIDFromContext(r.Context())

	profileUserPosts, err := h.postService.GetPostByUser(r.Context(), userID, currentUserID)
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	// TODO: here get liked post by user service will be called
	// profileUserlikedPosts = nil
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, apperrors.ErrNotFound.Error(), http.StatusNotFound)
			return
		}

		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	data := &UserPageData{
		Title:       "User",
		User:        viewer,
		ProfileUser: profileUser,
		Categories:  categories,
		Posts:       profileUserPosts,
		LikedPosts:  nil,
	}
	switch r.URL.Query().Get("profile_picture_status") {
	case "success":
		data.ProfilePictureNotification = "Successfully updated your profile picture."
	case "error":
		data.ProfilePictureNotification = "Failed to update profile picture."
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
	redirectWithError := func() {
		http.Redirect(w, r, fmt.Sprintf("/users/%d?profile_picture_status=error", user.ID), http.StatusSeeOther)
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	file, header, err := r.FormFile("profile_picture")
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			redirectWithError()
			return
		}

		redirectWithError()
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		redirectWithError()
		return
	}

	// Check the file contents instead of trusting its extension
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		redirectWithError()
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
		redirectWithError()
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		apperrors.Log(err)
		redirectWithError()
		return
	}

	profileDir := filepath.Join("internal", "web", "static", "img", "profile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		apperrors.Log(err)
		redirectWithError()
		return
	}

	filename := uuid.New().String() + ext
	picturePath := filepath.Join(profileDir, filename)

	destination, err := os.Create(picturePath)
	if err != nil {
		apperrors.Log(err)
		redirectWithError()
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
		redirectWithError()
		return
	}

	if err := destination.Close(); err != nil {
		apperrors.Log(err)
		redirectWithError()
		return
	}

	pictureURL := "/static/img/profile/" + url.PathEscape(filename)

	if err := h.userService.UpdateProfilePicture(r.Context(), user.ID, pictureURL); err != nil {
		apperrors.Log(err)
		redirectWithError()
		return
	}

	saved = true

	// Redirect the POST to a GET so the browser reloads the updated profile
	// without resubmitting the upload when the page is refreshed.
	http.Redirect(w, r, fmt.Sprintf("/users/%d?profile_picture_status=success", user.ID), http.StatusSeeOther)
}
