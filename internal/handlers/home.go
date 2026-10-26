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
	Title       string
	User        *models.User
	Categories  []models.Category
	Posts       []models.Post
	SearchQuery string
	Popular     []models.PopularPost
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
	data := &HomePageData{
		Title:       "Home",
		User:        user,
		Categories:  categories,
		Posts:       posts,
		SearchQuery: searchQuery,
		Popular:     popular,
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
