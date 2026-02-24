# Recipe Backend API

Go REST API backend for the Marellis Recipe Website.

## Features

- RESTful API with CRUD operations for recipes
- Category management and filtering
- Full-text search across recipes
- SQLite database with JSON column support
- CORS enabled for GitHub Pages frontend
- Graceful shutdown
- Health check endpoint

## Tech Stack

- **Language**: Go 1.21+
- **Router**: Chi v5
- **Database**: SQLite with sqlx
- **Validation**: go-playground/validator

## Project Structure

```
recipe-backend/
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # CORS, logging middleware
│   ├── models/          # Data models
│   ├── repository/      # Database operations
│   └── router/          # Route definitions
├── migrations/          # SQL schema migrations
└── go.mod
```

## Prerequisites

- Go 1.21 or higher
- SQLite support (included with modernc.org/sqlite - pure Go)

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and update values:
   ```bash
   cp .env.example .env
   ```
3. Install dependencies:
   ```bash
   make deps
   ```

## Running Locally

```bash
# Run in development mode
make run

# Or directly with go
go run cmd/api/main.go
```

The server will start on `http://localhost:8080` (or the port specified in `.env`).

## API Endpoints

### Recipes

- `GET /api/v1/recipes` - List all recipes (supports `?category=` and `?q=` query params)
- `GET /api/v1/recipes/:id` - Get single recipe
- `POST /api/v1/recipes` - Create new recipe
- `PUT /api/v1/recipes/:id` - Update recipe
- `DELETE /api/v1/recipes/:id` - Delete recipe

### Categories

- `GET /api/v1/categories` - Get all unique categories

### Health Check

- `GET /health` - Health check endpoint

## Example Requests

### Create Recipe

```bash
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Classic Pancakes",
    "description": "Fluffy pancakes",
    "categories": ["breakfast", "vegetarian"],
    "ingredients": ["1 cup flour", "2 eggs", "1 cup milk"],
    "steps": ["Mix ingredients", "Cook on griddle"],
    "links": []
  }'
```

### Get All Recipes

```bash
curl http://localhost:8080/api/v1/recipes
```

### Search Recipes

```bash
curl "http://localhost:8080/api/v1/recipes?q=pancake"
```

### Filter by Category

```bash
curl "http://localhost:8080/api/v1/recipes?category=breakfast"
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Building

```bash
# Build binary
make build

# Run binary
./bin/api
```

## Deployment

### Fly.io

1. Install flyctl: `https://fly.io/docs/hands-on/install-flyctl/`
2. Login: `flyctl auth login`
3. Initialize app: `flyctl launch`
4. Deploy: `flyctl deploy`

See `Dockerfile` for containerization setup.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `APP_ENV` | Environment (development/production) | `development` |
| `DATABASE_PATH` | SQLite database file path | `./recipes.db` |
| `CORS_ORIGIN` | Allowed CORS origins (comma-separated) | `http://localhost:8080` |

## License

Personal hobby project for family recipe sharing.
