package handlers

import (
	"net/http"

	"recipe-backend/internal/repository"
	"recipe-backend/internal/respond"
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
		respond.InternalError(w, "Failed to fetch categories", err)
		return
	}

	respond.JSON(w, http.StatusOK, categories)
}
