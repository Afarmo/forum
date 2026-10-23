package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

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
}

func (h *HomeHandler) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	categoryID := 0
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		var convErr error
		categoryID, convErr = strconv.Atoi(categoryIDStr)
		if convErr != nil {
			http.Error(w, "invalid category_id", http.StatusBadRequest)
			return
		}
	}

	searchQuery := strings.TrimSpace(r.URL.Query().Get("query"))
	var posts []models.Post
	if searchQuery != "" {
		posts, err = h.postService.SearchPosts(r.Context(), searchQuery)
		if err != nil {
			log.Println("search posts error:", err)
			posts = []models.Post{}
		}
	} else {
		posts, err = h.postService.GetAllPosts(r.Context(), categoryID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if posts == nil {
		posts = []models.Post{}
	}

	user := middleware.UserFromContext(r.Context())
	data := &HomePageData{
		Title:       "Home",
		User:        user,
		Categories:  categories,
		Posts:       posts,
		SearchQuery: searchQuery,
	}

	var buf bytes.Buffer

	if err := h.tmpl.ExecuteTemplate(&buf, "home", data); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Println(err)
		return
	}
}
