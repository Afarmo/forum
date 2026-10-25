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
	Title       string
	User        *models.User
	ProfileUser *models.User
	Categories  []models.Category
	Posts       []models.Post
	LikedPosts  []models.Post
	SearchQuery string
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
	profileUserlikedPosts, err := h.postService.GetPostByUser(r.Context(), userID, currentUserID)
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
		LikedPosts:  profileUserlikedPosts,
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

	file, header, err := r.FormFile("profile_picture")
	if err != nil {
		http.Error(w, "failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err = os.MkdirAll("internal/web/static/img/profile", 0755); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	filename := filepath.Base(header.Filename)
	picturePath := filepath.Join("internal", "web", "static", "img", "profile", filename)
	destination, err := os.Create(picturePath)
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}
	defer destination.Close()

	if _, err = io.Copy(destination, file); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	user := middleware.UserFromContext(r.Context())

	pictureURL := "/static/img/profile/" + url.PathEscape(filename)
	if err = h.userService.UpdateProfilePicture(r.Context(), user.ID, pictureURL); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	// Redirect the POST to a GET so the browser reloads the updated profile
	// without resubmitting the upload when the page is refreshed.
	http.Redirect(w, r, fmt.Sprintf("/users/%d", user.ID), http.StatusSeeOther)
}
