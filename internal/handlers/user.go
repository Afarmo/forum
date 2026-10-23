package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	// "strconv"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/middleware"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type UserHandler struct {
	categoryService *service.CategoryService
	postService     *service.PostService
	userService     *service.UserService
	tmpl            *template.Template
}

func NewUserHandler(userService *service.UserService, categoryService *service.CategoryService, postService *service.PostService, tmpl *template.Template) *UserHandler {
	return &UserHandler{
		userService:     userService,
		categoryService: categoryService,
		postService:     postService,
		tmpl:            tmpl,
	}
}

type UserPageData struct {
	Title      string
	User       *models.User
	Categories []models.Category
	Posts      []models.Post
	LikedPosts []models.Post
	SearchQuery string
}

func (h *UserHandler) FindUserById(w http.ResponseWriter, r *http.Request) {

	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// categoryID := 0
	// if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
	// 	var convErr error
	// 	categoryID, convErr = strconv.Atoi(categoryIDStr)
	// 	if convErr != nil {
	// 		http.Error(w, "invalid category_id", http.StatusBadRequest)
	// 		return
	// 	}
	// }
	user := middleware.UserFromContext(r.Context())

	myPosts, err := h.postService.GetPostByUser(r.Context(), user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// TODO: here get liked post by user service will be called
	likedPost, err := h.postService.GetPostByUser(r.Context(), user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		http.Error(w, apperrors.ErrNotFound.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := &UserPageData{
		Title:      "User",
		User:       user,
		Categories: categories,
		Posts:      myPosts,
		LikedPosts: likedPost,
	}

	var buf bytes.Buffer

	if err := h.tmpl.ExecuteTemplate(&buf, "user", data); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Println(err)
		return
	}
}

func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("profile_picture")
	if err != nil {
		log.Println("upload failed: ", err)
		http.Error(w, "failed to get uploaded file", http.StatusBadRequest) // TODO
		return
	}
	defer file.Close()

	if err = os.MkdirAll("internal/web/static/img/profile", 0755); err != nil {
		http.Error(w, "failed to create upload directory", http.StatusInternalServerError) // TODO
		return
	}

	filename := filepath.Base(header.Filename)
	picturePath := filepath.Join("internal", "web", "static", "img", "profile", filename)
	destination, err := os.Create(picturePath)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError) // TODO
		return
	}
	defer destination.Close()

	if _, err = io.Copy(destination, file); err != nil {
		http.Error(w, "failed to save uploaded file", http.StatusInternalServerError) // TODO
		return
	}
	user := middleware.UserFromContext(r.Context())

	pictureURL := "/static/img/profile/" + url.PathEscape(filename)
	if err = h.userService.UpdateProfilePicture(r.Context(), user.ID, pictureURL); err != nil {
		http.Error(w, "failed to update profile picture", http.StatusInternalServerError)
		return
	}
	// Redirect the POST to a GET so the browser reloads the updated profile
	// without resubmitting the upload when the page is refreshed.
	http.Redirect(w, r, fmt.Sprintf("/users/%d", user.ID), http.StatusSeeOther)
}
