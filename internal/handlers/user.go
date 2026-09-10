package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
	"github.com/Afarmo/forum/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserRepository(service *service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest) // TODO
		return
	}

	if err := h.service.CreateUser(ctx, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // TODO
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) FindUserByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := strings.TrimSpace(r.URL.Query().Get("email"))
	user, err := h.service.FindUserByEmail(ctx, email)
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
		http.Error(w, "failed to get uploaded file", http.StatusBadRequest) // TODO
		return
	}
	defer file.Close()

	if err = os.MkdirAll("uploads/profiles", 0755); err != nil {
		http.Error(w, "failed to create upload directory", http.StatusInternalServerError) // TODO
		return
	}

	picturePath := "uploads/profiles/" + header.Filename
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
