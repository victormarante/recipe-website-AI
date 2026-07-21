# Backend API

Go REST API for the Marellis Recipe Website.

## Responsibilities

- Load environment configuration
- Open and migrate the SQLite database
- Serve JWT-protected recipe, category, and image endpoints
- Store recipe arrays as JSON text in SQLite
- Optionally upload/delete recipe images in Cloudflare R2-compatible storage

## Structure

```text
backend/
├── cmd/api/main.go
├── internal/config/
├── internal/database/
├── internal/handlers/
├── internal/middleware/
├── internal/models/
├── internal/repository/
├── internal/router/
├── migrations/
├── go.mod
└── Makefile
```

## Configuration

Required:

- `AUTH_USERNAME`
- `AUTH_PASSWORD` or `AUTH_PASSWORD_HASH`
- `JWT_SECRET`

Common:

- `PORT` defaults to `8080`
- `APP_ENV` defaults to `development`
- `DATABASE_PATH` defaults to `./recipes.db`
- `CORS_ORIGIN` defaults to `http://localhost:8080`

Optional R2 image storage:

- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET_NAME`
- `R2_PUBLIC_URL`

The app initializes image storage only when all R2 variables are present. A partial R2 configuration fails startup. Leave all R2 variables empty to disable image endpoints.

For production, prefer `AUTH_PASSWORD_HASH` with a bcrypt hash. `AUTH_PASSWORD` remains supported for local development and existing deployments.

## Run Locally

```bash
cp .env.example .env
go mod download
go run cmd/api/main.go
```

Or:

```bash
make run
```

## Validation

```bash
gofmt -w ./...
go build ./...
go test ./...
go vet ./...
```

The test suite covers repository behavior, migrations, authentication middleware, handlers, and router behavior.

## API

Base path: `/api/v1`

Public:

- `GET /health`
- `GET /ready`
- `POST /api/v1/auth/login`

Protected by `Authorization: Bearer <token>`:

- `GET /api/v1/recipes`
- `POST /api/v1/recipes`
- `GET /api/v1/recipes/{id}`
- `PUT /api/v1/recipes/{id}`
- `DELETE /api/v1/recipes/{id}`
- `POST /api/v1/recipes/{id}/image`
- `DELETE /api/v1/recipes/{id}/image`
- `GET /api/v1/categories`

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"change-me"}'
```

Response:

```json
{"token":"<jwt>"}
```

The returned JWT is valid for 90 days.

### List Recipes

```bash
curl http://localhost:8080/api/v1/recipes \
  -H "Authorization: Bearer $TOKEN"
```

Optional query parameters:

- `category`
- `q`

Search uses SQL `LIKE` matching across title, description, categories, and ingredients. It is not SQLite FTS.

### Create Recipe

```bash
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Classic Pancakes",
    "description": "Fluffy pancakes",
    "categories": ["breakfast", "vegetarian"],
    "ingredients": ["1 cup flour", "2 eggs", "1 cup milk"],
    "steps": ["Mix ingredients", "Cook on griddle"],
    "links": [],
    "oven_temperature": null
  }'
```

### Upload Recipe Image

Requires R2 configuration and an existing recipe.

```bash
curl -X POST http://localhost:8080/api/v1/recipes/1/image \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@photo.jpg"
```

The request is limited to 5 MB. Detected content type must be `image/jpeg`, `image/png`, or `image/webp`.

Images are stored under `recipes/{id}` in the configured bucket. Uploading a new image for the same recipe overwrites the previous object at that key. The backend does not currently set custom cache-control headers.

### Delete Recipe Image

```bash
curl -X DELETE http://localhost:8080/api/v1/recipes/1/image \
  -H "Authorization: Bearer $TOKEN"
```

## Database

`migrations/001_create_tables.sql` creates the base `recipes` table and indexes. `migrations/002_add_recipe_metadata.sql` adds `oven_temperature` and `image_url`.

Startup applies ordered SQL migrations once and records them in `schema_migrations`. Fresh databases reach the latest schema, and old databases are upgraded without resetting recipe data.

## Deployment

The active Fly.io deployment path uses the root `fly.toml` and root `Dockerfile`.
