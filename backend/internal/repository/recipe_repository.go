package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"recipe-backend/internal/models"
)

// RecipeRepository handles database operations for recipes
type RecipeRepository struct {
	db *sqlx.DB
}

// NewRecipeRepository creates a new RecipeRepository instance
func NewRecipeRepository(db *sqlx.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// dbRecipe represents a recipe as stored in the database (with JSON fields as strings)
type dbRecipe struct {
	ID          int64  `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Categories  string `db:"categories"`
	Ingredients string `db:"ingredients"`
	Steps       string `db:"steps"`
	Links       string `db:"links"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

// Create inserts a new recipe into the database
func (r *RecipeRepository) Create(req models.CreateRecipeRequest) (*models.Recipe, error) {
	// Marshal arrays to JSON
	categoriesJSON, _ := json.Marshal(req.Categories)
	ingredientsJSON, _ := json.Marshal(req.Ingredients)
	stepsJSON, _ := json.Marshal(req.Steps)
	linksJSON, _ := json.Marshal(req.Links)

	query := `
		INSERT INTO recipes (title, description, categories, ingredients, steps, links)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		req.Title,
		req.Description,
		string(categoriesJSON),
		string(ingredientsJSON),
		string(stepsJSON),
		string(linksJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Fetch and return the created recipe
	return r.FindByID(id)
}

// FindAll retrieves all recipes with optional filtering
func (r *RecipeRepository) FindAll(category, searchQuery string) ([]models.Recipe, error) {
	query := `SELECT id, title, description, categories, ingredients, steps, links, created_at, updated_at FROM recipes WHERE 1=1`
	args := []interface{}{}

	// Filter by category if provided
	if category != "" {
		query += ` AND (categories LIKE ? OR json_extract(categories, '$') LIKE ?)`
		categoryPattern := fmt.Sprintf("%%\"%s\"%%", strings.ToLower(category))
		args = append(args, categoryPattern, categoryPattern)
	}

	// Search if query provided
	if searchQuery != "" {
		query += ` AND (
			title LIKE ? OR 
			description LIKE ? OR 
			categories LIKE ? OR 
			ingredients LIKE ?
		)`
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	query += ` ORDER BY created_at DESC`

	var dbRecipes []dbRecipe
	if err := r.db.Select(&dbRecipes, query, args...); err != nil {
		return nil, fmt.Errorf("failed to query recipes: %w", err)
	}

	// Convert dbRecipes to models.Recipe
	recipes := make([]models.Recipe, 0, len(dbRecipes))
	for _, dbr := range dbRecipes {
		recipe, err := r.dbRecipeToModel(dbr)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *recipe)
	}

	return recipes, nil
}

// FindByID retrieves a single recipe by ID
func (r *RecipeRepository) FindByID(id int64) (*models.Recipe, error) {
	query := `SELECT id, title, description, categories, ingredients, steps, links, created_at, updated_at FROM recipes WHERE id = ?`

	var dbr dbRecipe
	if err := r.db.Get(&dbr, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to query recipe: %w", err)
	}

	return r.dbRecipeToModel(dbr)
}

// Update updates an existing recipe
func (r *RecipeRepository) Update(id int64, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	// Check if recipe exists
	if _, err := r.FindByID(id); err != nil {
		return nil, err
	}

	// Marshal arrays to JSON
	categoriesJSON, _ := json.Marshal(req.Categories)
	ingredientsJSON, _ := json.Marshal(req.Ingredients)
	stepsJSON, _ := json.Marshal(req.Steps)
	linksJSON, _ := json.Marshal(req.Links)

	query := `
		UPDATE recipes 
		SET title = ?, description = ?, categories = ?, ingredients = ?, steps = ?, links = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		req.Title,
		req.Description,
		string(categoriesJSON),
		string(ingredientsJSON),
		string(stepsJSON),
		string(linksJSON),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	// Fetch and return the updated recipe
	return r.FindByID(id)
}

// Delete removes a recipe from the database
func (r *RecipeRepository) Delete(id int64) error {
	query := `DELETE FROM recipes WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recipe not found")
	}

	return nil
}

// GetAllCategories retrieves all unique categories from all recipes
func (r *RecipeRepository) GetAllCategories() ([]string, error) {
	query := `SELECT DISTINCT categories FROM recipes`

	var categoriesJSONList []string
	if err := r.db.Select(&categoriesJSONList, query); err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}

	// Extract unique categories from all recipes
	categorySet := make(map[string]bool)
	for _, categoriesJSON := range categoriesJSONList {
		var categories []string
		if err := json.Unmarshal([]byte(categoriesJSON), &categories); err != nil {
			continue
		}
		for _, cat := range categories {
			categorySet[strings.ToLower(strings.TrimSpace(cat))] = true
		}
	}

	// Convert set to sorted slice
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	return categories, nil
}

// dbRecipeToModel converts a database recipe to a model recipe
func (r *RecipeRepository) dbRecipeToModel(dbr dbRecipe) (*models.Recipe, error) {
	recipe := &models.Recipe{
		ID:          dbr.ID,
		Title:       dbr.Title,
		Description: dbr.Description,
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(dbr.Categories), &recipe.Categories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories: %w", err)
	}
	if err := json.Unmarshal([]byte(dbr.Ingredients), &recipe.Ingredients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ingredients: %w", err)
	}
	if err := json.Unmarshal([]byte(dbr.Steps), &recipe.Steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}

	// Links might be null in database
	if dbr.Links != "" && dbr.Links != "null" {
		if err := json.Unmarshal([]byte(dbr.Links), &recipe.Links); err != nil {
			return nil, fmt.Errorf("failed to unmarshal links: %w", err)
		}
	} else {
		recipe.Links = []models.Link{}
	}

	return recipe, nil
}
