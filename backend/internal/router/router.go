package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"recipe-backend/internal/config"
	"recipe-backend/internal/handlers"
	customMiddleware "recipe-backend/internal/middleware"
)

// New creates and configures the application router
func New(
	recipeHandler *handlers.RecipeHandler,
	categoryHandler *handlers.CategoryHandler,
	authHandler *handlers.AuthHandler,
	cfg *config.Config,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.Logger)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public: login
		r.Post("/auth/login", authHandler.Login)

		// Protected routes — require valid JWT
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.JWTAuth(cfg.JWTSecret))

			r.Get("/recipes", recipeHandler.GetRecipes)
			r.Post("/recipes", recipeHandler.CreateRecipe)
			r.Get("/recipes/{id}", recipeHandler.GetRecipe)
			r.Put("/recipes/{id}", recipeHandler.UpdateRecipe)
			r.Delete("/recipes/{id}", recipeHandler.DeleteRecipe)

			r.Get("/categories", categoryHandler.GetCategories)
		})
	})

	return r
}
