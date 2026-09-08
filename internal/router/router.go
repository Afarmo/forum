package router

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/Afarmo/forum/internal/handlers"
	"github.com/Afarmo/forum/internal/repository"
	"github.com/Afarmo/forum/internal/service"
)

func NewRouter(tmpl *template.Template, db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	repo := repository.NewUserRepository(db)
	userService := service.NewUserService(repo)
	userHandler := handlers.NewUserRepository(userService)

	mux.HandleFunc("GET /", handlers.HomeHandler(tmpl))
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
