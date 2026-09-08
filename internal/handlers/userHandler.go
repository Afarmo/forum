package handlers

import (
	"encoding/json"
	"net/http"

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
	cx := r.Context()
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	err = h.service.CreateUser(cx, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
