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

	"github.com/Afarmo/forum/internal/database"
	"github.com/Afarmo/forum/internal/handlers"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/repository"
	"github.com/Afarmo/forum/internal/router"
	"github.com/Afarmo/forum/internal/service"
	_ "github.com/mattn/go-sqlite3"
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


	// services
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo)
	authService := service.NewAuthService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)


	// handlers
	userHandler := handlers.NewUserHandler(userService, tmpl)
	postHandler := handlers.NewPostHandler(postService, tmpl)
	authHandler := handlers.NewAuthHandler(authService, tmpl)
	homeHandler := handlers.NewHomeHandler(tmpl, categoryService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)


	mux := router.NewRouter(homeHandler, userHandler, postHandler, authHandler, categoryHandler)
	handler := middleware.Recover(middleware.Logger(mux))
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
