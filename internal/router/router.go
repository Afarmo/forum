package router

import (
	"net/http"

	"github.com/Afarmo/forum/internal/handlers"
)

func NewRouter(homeHandler *handlers.HomeHandler,
	userHandler *handlers.UserHandler,
	postHandler *handlers.PostHandler,
	authHandler *handlers.AuthHandler,
	categoryHandler *handlers.CategoryHandler,
	commentHandler *handlers.CommentHandler,
	reactionHandler *handlers.ReactionHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", homeHandler.HomePageHandler)
	mux.HandleFunc("POST /register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /login", authHandler.LoginHandler)

	mux.HandleFunc("GET /users/{id}", userHandler.FindUserById)
	mux.HandleFunc("POST /users/profile_picture", userHandler.UploadProfilePicture)

	mux.HandleFunc("POST /writePosts", postHandler.CreatePost)
	mux.HandleFunc("GET /posts/{id}", postHandler.CreatePostPage)
	mux.HandleFunc("PATCH /posts/{id}", postHandler.UpdatePost)
	mux.HandleFunc("DELETE /posts/{id}", postHandler.DeletePost)
	mux.HandleFunc("GET /posts", postHandler.GetAllPosts)
	mux.HandleFunc("GET /users/{id}/posts", postHandler.GetPostByUser)
	mux.HandleFunc("GET /categories", categoryHandler.GetAllCategories)
	mux.HandleFunc("GET /users/{id}/categories", categoryHandler.GetCategoryByUser)
	mux.HandleFunc("POST /posts/{id}/comments", commentHandler.CreateComment)
	mux.HandleFunc("GET /posts/{id}/comments", commentHandler.GetCommentsByPost)
	mux.HandleFunc("GET /search", postHandler.SearchPosts)

	mux.HandleFunc("POST /posts/{id}/reaction", reactionHandler.ReactToPost)
	mux.HandleFunc("POST /posts/{postID}/comments/{commentID}/reaction", reactionHandler.ReactToComment)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
