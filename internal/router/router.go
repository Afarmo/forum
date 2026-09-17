package router

import (
	"net/http"

	"github.com/Afarmo/forum/internal/handlers"
)

func NewRouter(homeHandler *handlers.HomeHandler,
	userHandler *handlers.UserHandler,
	postHandler *handlers.PostHandler,
	authHandler *handlers.AuthHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", homeHandler.HomePageHandler)
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users/email", userHandler.FindUserByEmail)
	mux.HandleFunc("GET /users/{id}", userHandler.FindUserById)
	mux.HandleFunc("POST /users/profile_picture", userHandler.UploadProfilePicture)

	mux.HandleFunc("POST /posts", postHandler.CreatePost)
	mux.HandleFunc("PATCH /posts/{id}", postHandler.UpdatePost)
	mux.HandleFunc("DELETE /posts/{id}", postHandler.DeletePost)
	mux.HandleFunc("GET /posts", postHandler.GetAllPosts)
	mux.HandleFunc("GET /users/{id}/posts", postHandler.GetPostByUser)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
