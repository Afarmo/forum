package router

import (
	"net/http"

	"github.com/Afarmo/forum/internal/app"
	"github.com/Afarmo/forum/internal/handlers"
	"github.com/Afarmo/forum/internal/repository"
	"github.com/Afarmo/forum/internal/service"
)

func NewRouter(a *app.Application) *http.ServeMux {
	mux := http.NewServeMux()
	repoUser := repository.NewUserRepository(a.DB)
	userService := service.NewUserService(repoUser)
	userHandler := handlers.NewUserHandler(userService)

	mux.HandleFunc("GET /", handlers.HomeHandler(a))
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users/email", userHandler.FindUserByEmail)
	mux.HandleFunc("GET /users/{id}", userHandler.FindUserById)
	mux.HandleFunc("POST /users/profile_picture", userHandler.UploadProfilePicture)

	repoPost := repository.NewPostRepository(a.DB)
	postService := service.NewPostService(repoPost)
	postHandler := handlers.NewPostHandler(postService)

	mux.HandleFunc("POST /posts", postHandler.CreatePost)
	mux.HandleFunc("GET /posts", postHandler.GetAllPosts)
	mux.HandleFunc("GET /users/{id}/posts", postHandler.GetPostByUser)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
