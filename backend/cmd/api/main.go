package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	migrationsDir := "migrations"
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	repo := repository.NewRecipeRepository(db)
	recipeHandler := handlers.NewRecipeHandler(repo, cfg)
	categoryHandler := handlers.NewCategoryHandler(repo)
	authHandler := handlers.NewAuthHandler(cfg)

	r := router.New(recipeHandler, categoryHandler, authHandler, cfg, db)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("error: graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
