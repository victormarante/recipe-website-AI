# Backend API

Go REST API for the Marellis Recipe Website.

## Responsibilities

- Load environment configuration
- Open and migrate the SQLite database
- Serve public read endpoints; recipe and image writes require an admin token from PIN login
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

Common:

- `PORT` defaults to `8080`
- `APP_ENV` defaults to `development`
- `DATABASE_PATH` defaults to `./recipes.db`
- `CORS_ORIGIN` defaults to `http://localhost:8080`

Admin access variables:

- `ADMIN_PIN` - numeric, at least 6 digits. Required when `APP_ENV=production`; in development it defaults to `123456`.
- `AUTH_SECRET` - random string used to sign tokens (e.g. `openssl rand -hex 32`). Required when `APP_ENV=production`.

Tokens are valid for 90 days and are stored in the browser's `localStorage`; changing `AUTH_SECRET` invalidates all of them.

Optional R2 image storage:

- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET_NAME`
- `R2_PUBLIC_URL`

The app initializes image storage only when all R2 variables are present. A partial R2 configuration fails startup. Leave all R2 variables empty to disable image endpoints.

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

The test suite covers repository behavior, migrations, handlers, and router behavior.

## API

Base path: `/api/v1`

Read endpoints are public. Write endpoints require an admin token (`Authorization: Bearer <token>`), obtained by posting the admin PIN:

- `GET /health`
- `GET /ready`
- `POST /api/v1/auth/login` (public; body `{"pin":"123456"}`, returns `{"token","expires_at"}`; 401 on a wrong PIN, 429 after 5 failures per 15 minutes per IP)
- `GET /api/v1/recipes`
- `GET /api/v1/recipes/{id}`
- `GET /api/v1/categories`
- `POST /api/v1/recipes` (admin)
- `PUT /api/v1/recipes/{id}` (admin)
- `DELETE /api/v1/recipes/{id}` (admin)
- `POST /api/v1/recipes/{id}/image` (admin)
- `DELETE /api/v1/recipes/{id}/image` (admin)

### List Recipes

```bash
curl http://localhost:8080/api/v1/recipes
```

Optional query parameters:

- `category`
- `q`

Search uses SQL `LIKE` matching across title, description, categories, and ingredients. It is not SQLite FTS.

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
    "links": [],
    "oven_temperature": null
  }'
```

Optional thumbnail crop fields (create and update; responses always include them, `null` meaning centred with no zoom): `thumb_x` and `thumb_y` (0-100, the part of the image shown in the list thumbnail) and `thumb_zoom` (0.05-5; 1 fits the image to cover the frame, lower values show the whole image). The frontend uses them to frame the image in the recipe list. `PUT` replaces them, so omitting them resets the crop.

Optional `oven_mode` (create and update; responses always include it, `null` when unspecified): `"fan"` (convection, Swedish "varmluft") or `"conventional"` (top and bottom heat, "över- och undervärme"); any other value is rejected with 400. `PUT` replaces it, so omitting it clears it. The API does not require it alongside `oven_temperature`; the frontend form does.

### Upload Recipe Image

Requires R2 configuration and an existing recipe.

```bash
curl -X POST http://localhost:8080/api/v1/recipes/1/image \
  -F "image=@photo.jpg"
```

The request is limited to 5 MB. Detected content type must be `image/jpeg`, `image/png`, or `image/webp`.

Images are stored under `recipes/{id}` in the configured bucket. Uploading a new image for the same recipe overwrites the previous object at that key. The backend does not currently set custom cache-control headers.

### Delete Recipe Image

```bash
curl -X DELETE http://localhost:8080/api/v1/recipes/1/image
```

## Database

`migrations/001_create_tables.sql` creates the base `recipes` table and indexes. `migrations/002_add_recipe_metadata.sql` adds `oven_temperature` and `image_url`. `migrations/003_add_thumbnail_crop.sql` adds `thumb_x`, `thumb_y` and `thumb_zoom`. `migrations/004_add_oven_mode.sql` adds `oven_mode`.

Startup applies ordered SQL migrations once and records them in `schema_migrations`. Fresh databases reach the latest schema, and old databases are upgraded without resetting recipe data.

## Deployment

The active Fly.io deployment path uses the root `fly.toml` and root `Dockerfile`.
