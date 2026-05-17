package main

import (
	"log"
	"net/http"
	"path/filepath"

	"recipe-backend/internal/config"
	"recipe-backend/internal/database"
	"recipe-backend/internal/handlers"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close(db)

	migrationPath := filepath.Join("migrations", "001_create_tables.sql")
	if err := database.RunMigrations(db, migrationPath); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	if err := database.AddColumnIfNotExists(db, "recipes", "oven_temperature", "INTEGER DEFAULT NULL"); err != nil {
		log.Fatalf("migration: %v", err)
	}

	if err := database.AddColumnIfNotExists(db, "recipes", "image_url", "TEXT"); err != nil {
		log.Fatalf("migration: %v", err)
	}

	repo := repository.NewRecipeRepository(db)
	recipeHandler := handlers.NewRecipeHandler(repo, cfg)
	categoryHandler := handlers.NewCategoryHandler(repo)
	authHandler := handlers.NewAuthHandler(cfg)

	r := router.New(recipeHandler, categoryHandler, authHandler, cfg)

	log.Printf("server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
