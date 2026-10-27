package router

import (
	"net/http"

	"github.com/Afarmo/forum/internal/handlers"
	"github.com/Afarmo/forum/internal/middleware"
)

func NewRouter(homeHandler *handlers.HomeHandler,
	userHandler *handlers.UserHandler,
	postHandler *handlers.PostHandler,
	authHandler *handlers.AuthHandler,
	categoryHandler *handlers.CategoryHandler,
	commentHandler *handlers.CommentHandler,
	reactionHandler *handlers.ReactionHandler,
	authMiddleware *middleware.AuthMiddleware,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /", homeHandler.HomePageHandler)
	mux.HandleFunc("POST /register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /login", authHandler.LoginHandler)
	mux.HandleFunc("GET /logout", authHandler.LogoutHandler)

	mux.HandleFunc("GET /posts", postHandler.GetAllPosts)
	mux.HandleFunc("GET /posts/{id}", postHandler.CreatePostPage)
	mux.HandleFunc("GET /posts/{id}/comments", commentHandler.GetCommentsByPost)
	mux.HandleFunc("GET /search", postHandler.SearchPosts)

	mux.HandleFunc("GET /users/{id}", userHandler.FindUserById)
	mux.HandleFunc("GET /users/{id}/posts", postHandler.GetPostByUser)
	mux.HandleFunc("GET /users/{id}/categories", categoryHandler.GetCategoryByUser)
	mux.HandleFunc("GET /categories", categoryHandler.GetAllCategories)

	// Protected routes
	protected := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, authMiddleware.RequireAuth(handler))
	}

	protected("POST /posts/{id}/update", postHandler.UpdatePost)
	protected("POST /posts/{id}/delete", postHandler.DeletePost)
	protected("POST /writePosts", postHandler.CreatePost)
	protected("POST /posts/{id}/comments", commentHandler.CreateComment)
	protected("POST /comments/{id}/delete", commentHandler.DeleteComment)
	protected("POST /posts/{id}/reaction", reactionHandler.ReactToPost)
	protected("POST /posts/{postID}/comments/{commentID}/reaction", reactionHandler.ReactToComment)
	protected("POST /users/profile_picture", userHandler.UploadProfilePicture)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	return mux
}
