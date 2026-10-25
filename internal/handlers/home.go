package handlers

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type HomeHandler struct {
	categoryService *service.CategoryService
	postService     *service.PostService
	tmpl            *template.Template
}

func NewHomeHandler(tmpl *template.Template, categoryService *service.CategoryService, postService *service.PostService) *HomeHandler {
	return &HomeHandler{
		tmpl:            tmpl,
		categoryService: categoryService,
		postService:     postService,
	}
}

type HomePageData struct {
	Title         string
	User          *models.User
	Categories    []models.Category
	Posts         []models.Post
	SearchQuery   string
	LoginError    string
	RegisterError string
	LoginSuccess  bool
}

func (h *HomeHandler) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		apperrors.RenderError(w, h.tmpl, http.StatusNotFound)
		return
	}

	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	var categoryID int
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		categoryID, err = strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category_id", http.StatusBadRequest)
			return
		}
	}

	searchQuery := strings.TrimSpace(r.URL.Query().Get("query"))
	user := middleware.UserFromContext(r.Context())
	currentUserID := middleware.UserIDFromContext(r.Context())

	var posts []models.Post
	if searchQuery != "" {
		posts, err = h.postService.SearchPosts(r.Context(), searchQuery, currentUserID)
		if err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}
	} else {
		posts, err = h.postService.GetAllPosts(r.Context(), categoryID, currentUserID)
		if err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}
	}

	if posts == nil {
		posts = []models.Post{}
	}

	data := &HomePageData{
		Title:       "Home",
		User:        user,
		Categories:  categories,
		Posts:       posts,
		SearchQuery: searchQuery,
	}
	if r.URL.Query().Get("login_error") == "1" {
		data.LoginError = "Invalid email or password"
	}
	data.LoginSuccess = r.URL.Query().Get("login_success") == "1"
	switch r.URL.Query().Get("register_error") {
	case "invalid_username":
		data.RegisterError = "Username must be between 3 and 20 characters."
	case "invalid_email":
		data.RegisterError = "Enter a valid email address."
	case "invalid_password":
		data.RegisterError = "Password must be between 8 and 64 characters."
	case "duplicate_email":
		data.RegisterError = "An account with this email already exists."
	case "duplicate_username":
		data.RegisterError = "That username is already taken."
	case "duplicate":
		data.RegisterError = "An account with that information already exists."
	}

	var buf bytes.Buffer

	if err := h.tmpl.ExecuteTemplate(&buf, "home", data); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		apperrors.Log(err)
		return
	}
}
