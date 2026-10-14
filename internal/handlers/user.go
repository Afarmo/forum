package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

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
}

func (h *UserHandler) FindUserById(w http.ResponseWriter, r *http.Request) {

	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	categoryID := 0
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		var convErr error
		categoryID, convErr = strconv.Atoi(categoryIDStr)
		if convErr != nil {
			http.Error(w, "invalid category_id", http.StatusBadRequest)
			return
		}
	}

	posts, err := h.postService.GetPostByUser(r.Context(), categoryID)
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

	user := middleware.UserFromContext(r.Context())
	data := &UserPageData{
		Title:      "User",
		User:       user,
		Categories: categories,
		Posts:      posts,
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

	picturePath := "internal/web/static/img/profile/" + header.Filename
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

	if err = h.userService.UpdateProfilePicture(r.Context(), user.ID, picturePath); err != nil {
		http.Error(w, "failed to update profile picture", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "file received successfully",
		"filename": header.Filename,
		"size":     header.Size,
	})
}
