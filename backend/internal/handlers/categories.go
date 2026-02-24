package handlers

import (
	"net/http"

	"recipe-backend/internal/repository"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	repo *repository.RecipeRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(repo *repository.RecipeRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// GetCategories handles GET /api/v1/categories
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAllCategories()
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Return as JSON array
	w.Write([]byte("["))
	for i, cat := range categories {
		if i > 0 {
			w.Write([]byte(","))
		}
		w.Write([]byte("\"" + cat + "\""))
	}
	w.Write([]byte("]"))
}
