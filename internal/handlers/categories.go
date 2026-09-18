package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Afarmo/forum/internal/service"
)

type CategoryHandler struct {
	service *service.CategoryService

}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categories, err := h.service.GetAllCategories(ctx)
	if err != nil {
		log.Println("Get categories error:", err)
		http.Error(w, "failed to get categories", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}

