package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/service"
)

type UserHandler struct {
	service *service.UserService
	tmpl    *template.Template
}

func NewUserHandler(service *service.UserService, tmpl *template.Template) *UserHandler {
	return &UserHandler{
		service: service,
		tmpl:    tmpl,
	}
}

func (h *UserHandler) FindUserById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	user, err := h.service.FindUserById(ctx, id)
	if err == sql.ErrNoRows {
		http.Error(w, apperrors.ErrNotFound.Error(), http.StatusNotFound) // TODO
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // TODO
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
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

	ctx := r.Context()
	userID := 1 // just for test... waiting for authentication to get the user id from the session
	if err = h.service.UpdateProfilePicture(ctx, userID, picturePath); err != nil {
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
