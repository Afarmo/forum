package router

import (
	"net/http"

	"github.com/Afarmo/forum/internal/app"
	"github.com/Afarmo/forum/internal/handlers"
)

func NewRouter(a *app.Application) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.HomeHandler(a))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
