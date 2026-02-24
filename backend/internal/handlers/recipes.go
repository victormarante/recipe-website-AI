package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"recipe-backend/internal/models"
	"recipe-backend/internal/repository"
)

// RecipeHandler handles HTTP requests for recipes
type RecipeHandler struct {
	repo     *repository.RecipeRepository
	validate *validator.Validate
}

// NewRecipeHandler creates a new RecipeHandler
func NewRecipeHandler(repo *repository.RecipeRepository) *RecipeHandler {
	return &RecipeHandler{
		repo:     repo,
		validate: validator.New(),
	}
}

// GetRecipes handles GET /api/v1/recipes
func (h *RecipeHandler) GetRecipes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	searchQuery := r.URL.Query().Get("q")

	recipes, err := h.repo.FindAll(category, searchQuery)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch recipes", err)
		return
	}

	h.sendJSON(w, http.StatusOK, recipes)
}

// GetRecipe handles GET /api/v1/recipes/:id
func (h *RecipeHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid recipe ID", err)
		return
	}

	recipe, err := h.repo.FindByID(id)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Recipe not found", err)
		return
	}

	h.sendJSON(w, http.StatusOK, recipe)
}

// CreateRecipe handles POST /api/v1/recipes
func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	recipe, err := h.repo.Create(req)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to create recipe", err)
		return
	}

	h.sendJSON(w, http.StatusCreated, recipe)
}

// UpdateRecipe handles PUT /api/v1/recipes/:id
func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid recipe ID", err)
		return
	}

	var req models.UpdateRecipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	recipe, err := h.repo.Update(id, req)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to update recipe", err)
		return
	}

	h.sendJSON(w, http.StatusOK, recipe)
}

// DeleteRecipe handles DELETE /api/v1/recipes/:id
func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid recipe ID", err)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to delete recipe", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// sendJSON sends a JSON response
func (h *RecipeHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError sends an error response
func (h *RecipeHandler) sendError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := models.ErrorResponse{
		Error:   message,
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(errResp)
}
