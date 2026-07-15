package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"recipe-backend/internal/config"
	"recipe-backend/internal/models"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/respond"
)

type objectStore interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// RecipeHandler handles HTTP requests for recipes
type RecipeHandler struct {
	repo     *repository.RecipeRepository
	validate *validator.Validate
	cfg      *config.Config
	s3Client objectStore
}

// NewRecipeHandler creates a new RecipeHandler
func NewRecipeHandler(repo *repository.RecipeRepository, cfg *config.Config) *RecipeHandler {
	h := &RecipeHandler{
		repo:     repo,
		validate: validator.New(),
		cfg:      cfg,
	}
	if cfg.R2AccountID != "" && cfg.R2AccessKey != "" && cfg.R2SecretKey != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion("auto"),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2AccessKey, cfg.R2SecretKey, "")),
		)
		if err == nil {
			h.s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID))
				o.UsePathStyle = true
			})
			log.Printf("info: R2 client initialized (bucket=%s)", cfg.R2BucketName)
		} else {
			log.Printf("error: failed to initialize R2 client: %v", err)
		}
	}
	return h
}

// GetRecipes handles GET /api/v1/recipes
func (h *RecipeHandler) GetRecipes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	searchQuery := r.URL.Query().Get("q")

	recipes, err := h.repo.FindAll(category, searchQuery)
	if err != nil {
		respond.InternalError(w, "Failed to fetch recipes", err)
		return
	}

	respond.JSON(w, http.StatusOK, recipes)
}

// GetRecipe handles GET /api/v1/recipes/:id
func (h *RecipeHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	recipe, err := h.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to fetch recipe", err)
		return
	}

	respond.JSON(w, http.StatusOK, recipe)
}

// CreateRecipe handles POST /api/v1/recipes
func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Validation failed")
		return
	}

	recipe, err := h.repo.Create(req)
	if err != nil {
		respond.InternalError(w, "Failed to create recipe", err)
		return
	}

	respond.JSON(w, http.StatusCreated, recipe)
}

// UpdateRecipe handles PUT /api/v1/recipes/:id
func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	var req models.UpdateRecipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Validation failed")
		return
	}

	recipe, err := h.repo.Update(id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to update recipe", err)
		return
	}

	respond.JSON(w, http.StatusOK, recipe)
}

// DeleteRecipe handles DELETE /api/v1/recipes/:id
func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to delete recipe", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
