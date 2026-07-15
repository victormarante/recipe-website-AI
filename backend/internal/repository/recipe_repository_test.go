package repository_test

import (
	"errors"
	"path/filepath"
	"testing"

	"recipe-backend/internal/database"
	"recipe-backend/internal/models"
	"recipe-backend/internal/repository"
)

func setupRepo(t *testing.T) *repository.RecipeRepository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "recipes.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { database.Close(db) })

	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return repository.NewRecipeRepository(db)
}

func validRecipe(title string) models.CreateRecipeRequest {
	return models.CreateRecipeRequest{
		Title:       title,
		Description: "A useful description",
		Categories:  []string{"Dinner", "Family"},
		Ingredients: []string{"salt", "pepper"},
		Steps:       []string{"mix", "cook"},
		Links: []models.Link{{
			Type:  "external",
			URL:   "https://example.com",
			Label: "source",
		}},
	}
}

func TestRecipeRepositoryCRUDSearchFilterImageAndLinks(t *testing.T) {
	repo := setupRepo(t)

	created, err := repo.Create(validRecipe("Tomato Pasta"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created ID")
	}
	if len(created.Links) != 1 || created.Links[0].URL != "https://example.com" {
		t.Fatalf("links were not serialized correctly: %#v", created.Links)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Title != "Tomato Pasta" {
		t.Fatalf("got title %q", found.Title)
	}

	searchResults, err := repo.FindAll("", "Tomato")
	if err != nil {
		t.Fatalf("FindAll search: %v", err)
	}
	if len(searchResults) != 1 {
		t.Fatalf("expected one search result, got %d", len(searchResults))
	}

	categoryResults, err := repo.FindAll("dinner", "")
	if err != nil {
		t.Fatalf("FindAll category: %v", err)
	}
	if len(categoryResults) != 1 {
		t.Fatalf("expected one category result, got %d", len(categoryResults))
	}

	temp := 200
	updated, err := repo.Update(created.ID, models.UpdateRecipeRequest{
		Title:           "Baked Pasta",
		Description:     "Updated",
		Categories:      []string{"Dinner"},
		Ingredients:     []string{"pasta"},
		Steps:           []string{"bake"},
		Links:           []models.Link{},
		OvenTemperature: &temp,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Baked Pasta" || updated.OvenTemperature == nil || *updated.OvenTemperature != 200 {
		t.Fatalf("unexpected updated recipe: %#v", updated)
	}

	imageURL := "https://images.example/recipes/1"
	if err := repo.UpdateImageURL(created.ID, &imageURL); err != nil {
		t.Fatalf("UpdateImageURL set: %v", err)
	}
	withImage, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID after image: %v", err)
	}
	if withImage.ImageURL == nil || *withImage.ImageURL != imageURL {
		t.Fatalf("unexpected image url: %#v", withImage.ImageURL)
	}
	if err := repo.UpdateImageURL(created.ID, nil); err != nil {
		t.Fatalf("UpdateImageURL clear: %v", err)
	}

	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(created.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestRecipeRepositoryMissingRecipe(t *testing.T) {
	repo := setupRepo(t)

	if _, err := repo.FindByID(999); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("FindByID expected ErrNotFound, got %v", err)
	}
	if err := repo.Delete(999); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("Delete expected ErrNotFound, got %v", err)
	}
	if err := repo.UpdateImageURL(999, nil); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("UpdateImageURL expected ErrNotFound, got %v", err)
	}
}
