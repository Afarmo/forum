package router

import (
	"html/template"
	"net/http"

	"github.com/Afarmo/forum/internal/handlers"
)

func NewRouter(tmpl *template.Template) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.HomeHandler(tmpl))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
