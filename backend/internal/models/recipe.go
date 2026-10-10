package models

import "time"

// Recipe represents a recipe in the database
type Recipe struct {
	ID              int64     `json:"id" db:"id"`
	Title           string    `json:"title" db:"title" validate:"required"`
	Description     string    `json:"description" db:"description"`
	Categories      []string  `json:"categories" db:"categories" validate:"required,min=1"`
	Tags            []string  `json:"tags" db:"tags"`
	Ingredients     []string  `json:"ingredients" db:"ingredients" validate:"required,min=1"`
	Steps           []string  `json:"steps" db:"steps" validate:"required,min=1"`
	Links           []Link    `json:"links" db:"links"`
	OvenTemperature *int      `json:"oven_temperature" db:"oven_temperature"`
	OvenMode        *string   `json:"oven_mode" db:"oven_mode"`
	ImageURL        *string   `json:"image_url" db:"image_url"`
	ThumbX          *float64  `json:"thumb_x" db:"thumb_x"`
	ThumbY          *float64  `json:"thumb_y" db:"thumb_y"`
	ThumbZoom       *float64  `json:"thumb_zoom" db:"thumb_zoom"`
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
	Tags            []string `json:"tags"`
	Ingredients     []string `json:"ingredients" validate:"required,min=1"`
	Steps           []string `json:"steps" validate:"required,min=1"`
	Links           []Link   `json:"links"`
	OvenTemperature *int     `json:"oven_temperature"`
	// "fan" (convection) or "conventional" (top and bottom heat). Nil means unspecified.
	OvenMode *string `json:"oven_mode" validate:"omitempty,oneof=fan conventional"`
	// Thumbnail crop for the list view: focal point as a percentage of the
	// image (0-100) and zoom factor (0.05-5, 1 = image just covers the
	// frame, below 1 shows the whole image with a backdrop). Nil means centred, no zoom.
	ThumbX    *float64 `json:"thumb_x" validate:"omitempty,gte=0,lte=100"`
	ThumbY    *float64 `json:"thumb_y" validate:"omitempty,gte=0,lte=100"`
	ThumbZoom *float64 `json:"thumb_zoom" validate:"omitempty,gte=0.05,lte=5"`
}

// UpdateRecipeRequest represents the request body for updating a recipe
type UpdateRecipeRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     string   `json:"description"`
	Categories      []string `json:"categories" validate:"required,min=1"`
	Tags            []string `json:"tags"`
	Ingredients     []string `json:"ingredients" validate:"required,min=1"`
	Steps           []string `json:"steps" validate:"required,min=1"`
	Links           []Link   `json:"links"`
	OvenTemperature *int     `json:"oven_temperature"`
	// "fan" (convection) or "conventional" (top and bottom heat). Nil means unspecified.
	OvenMode *string `json:"oven_mode" validate:"omitempty,oneof=fan conventional"`
	// Thumbnail crop for the list view: focal point as a percentage of the
	// image (0-100) and zoom factor (0.05-5, 1 = image just covers the
	// frame, below 1 shows the whole image with a backdrop). Nil means centred, no zoom.
	ThumbX    *float64 `json:"thumb_x" validate:"omitempty,gte=0,lte=100"`
	ThumbY    *float64 `json:"thumb_y" validate:"omitempty,gte=0,lte=100"`
	ThumbZoom *float64 `json:"thumb_zoom" validate:"omitempty,gte=0.05,lte=5"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
