package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/Afarmo/forum/internal/service"
)

type HomeHandler struct {
	service         *service.AuthService
	categoryService *service.CategoryService
	tmpl            *template.Template
}

func NewHomeHandler(tmpl *template.Template, categoryService *service.CategoryService) *HomeHandler {
	return &HomeHandler{
		tmpl:            tmpl,
		categoryService: categoryService,
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

	var buf bytes.Buffer

	data := map[string]any{
		"Title":      "Home",
		"Categories": categories,
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
