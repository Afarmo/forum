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
	reactionService *service.ReactionService
	tmpl            *template.Template
}

func NewHomeHandler(categoryService *service.CategoryService,
	postService *service.PostService,
	reactionService *service.ReactionService,
	tmpl *template.Template,
) *HomeHandler {
	return &HomeHandler{
		categoryService: categoryService,
		postService:     postService,
		reactionService: reactionService,
		tmpl:            tmpl,
	}
}

type HomePageData struct {
	Title       string
	User        *models.User
	Categories  []models.Category
	Posts       []models.Post
	SearchQuery string
	Popular     []models.PopularPost
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

	var posts []models.Post
	if searchQuery != "" {
		posts, err = h.postService.SearchPosts(r.Context(), searchQuery)
		if err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}
	} else {
		posts, err = h.postService.GetAllPosts(r.Context(), categoryID)
		if err != nil {
			apperrors.Log(err)
			apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
			return
		}
	}

	if posts == nil {
		posts = []models.Post{}
	}

	var popular []models.PopularPost
	if searchQuery == "" && categoryID == 0 {
		popular, err = h.postService.GetPopularThisWeek(r.Context())
		if err != nil {
			apperrors.Log(err)
			popular = nil
		}
	}

	user := middleware.UserFromContext(r.Context())

	var userID int
	if user != nil {
		userID = user.ID
	}

	if err := h.reactionService.PopulatePostStats(r.Context(), posts, userID); err != nil {
		apperrors.Log(err)
		apperrors.RenderError(w, h.tmpl, http.StatusInternalServerError)
		return
	}

	data := &HomePageData{
		Title:       "Home",
		User:        user,
		Categories:  categories,
		Posts:       posts,
		SearchQuery: searchQuery,
		Popular:     popular,
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
