package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/Afarmo/forum/internal/service"
)

type HomeHandler struct {
	service         *service.AuthService
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

	posts, err := h.postService.GetAllPosts(r.Context(), categoryID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer

	data := map[string]any{
		"Title":      "Home",
		"Categories": categories,
		"Posts":      posts,
	}

	if err := h.tmpl.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Println(err)
		return
	}
}
