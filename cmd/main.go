package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Afarmo/forum/internal/database"
	"github.com/Afarmo/forum/internal/handlers"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/repository"
	"github.com/Afarmo/forum/internal/router"
	"github.com/Afarmo/forum/internal/service"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.InitializeSchema(db); err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.New("").ParseGlob("internal/web/templates/*.html"))
	template.Must(tmpl.ParseGlob("internal/web/templates/partials/*.html"))

	// repositories
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// services
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo)
	authService := service.NewAuthService(userRepo, sessionRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	commentService := service.NewCommentService(commentRepo)

	// handlers
	userHandler := handlers.NewUserHandler(userService, categoryService, postService, tmpl)
	postHandler := handlers.NewPostHandler(postService, tmpl, commentService)
	authHandler := handlers.NewAuthHandler(authService, tmpl)
	homeHandler := handlers.NewHomeHandler(tmpl, categoryService, postService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	commentHandler := handlers.NewCommentHandler(commentService)

	mux := router.NewRouter(homeHandler, userHandler, postHandler, authHandler, categoryHandler, commentHandler)

	// middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)
	handler := middleware.Recover(
		middleware.Logger(
			authMiddleware.OptionalAuth(mux),
		),
	)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("\033[96m  [STARTUP]\033[0m  Listening on http://localhost:8080")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("\033[31m  [PANIC]\033[0m    Server error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("\033[35m[SHUTDOWN]\033[0m Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("\033[31m  [PANIC]\033[0m Forced shutdown:", err)
	}

	log.Println("\033[35m  [SHUTDOWN]\033[0m Server stopped")
}
