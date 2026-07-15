package models

import "time"

// Recipe represents a recipe in the database
type Recipe struct {
	ID              int64     `json:"id" db:"id"`
	Title           string    `json:"title" db:"title" validate:"required"`
	Description     string    `json:"description" db:"description"`
	Categories      []string  `json:"categories" db:"categories" validate:"required,min=1"`
	Ingredients     []string  `json:"ingredients" db:"ingredients" validate:"required,min=1"`
	Steps           []string  `json:"steps" db:"steps" validate:"required,min=1"`
	Links           []Link    `json:"links" db:"links"`
	OvenTemperature *int      `json:"oven_temperature" db:"oven_temperature"`
	ImageURL        *string   `json:"image_url" db:"image_url"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Link represents an external or recipe link
type Link struct {
	Type           string `json:"type" validate:"required,oneof=external recipe"`
	URL            string `json:"url,omitempty"`
	Label          string `json:"label,omitempty"`
	LinkedRecipeID *int64 `json:"linked_recipe_id,omitempty"`
}

// CreateRecipeRequest represents the request body for creating a recipe
type CreateRecipeRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     string   `json:"description"`
	Categories      []string `json:"categories" validate:"required,min=1"`
	Ingredients     []string `json:"ingredients" validate:"required,min=1"`
	Steps           []string `json:"steps" validate:"required,min=1"`
	Links           []Link   `json:"links"`
	OvenTemperature *int     `json:"oven_temperature"`
}

// UpdateRecipeRequest represents the request body for updating a recipe
type UpdateRecipeRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     string   `json:"description"`
	Categories      []string `json:"categories" validate:"required,min=1"`
	Ingredients     []string `json:"ingredients" validate:"required,min=1"`
	Steps           []string `json:"steps" validate:"required,min=1"`
	Links           []Link   `json:"links"`
	OvenTemperature *int     `json:"oven_temperature"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
