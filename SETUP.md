# Local Setup

This guide covers running the Marellis frontend and backend locally.

## Prerequisites

- Go compatible with the module in `backend/go.mod` (`go 1.24` at the time of writing)
- A local static file server for the frontend, such as Python 3
- Optional: `make`
- Optional: Fly CLI for deployment work

## Backend

```bash
cd backend
cp .env.example .env
```

Edit `.env` before starting the server if the defaults do not fit your setup.

```env
PORT=8080
APP_ENV=development
DATABASE_PATH=./recipes.db
CORS_ORIGIN=http://localhost:8080,http://localhost:8000
```

Install dependencies and run:

```bash
go mod download
go run cmd/api/main.go
```

The backend listens on `http://localhost:8080` by default.

Health check:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Frontend

```bash
cd frontend
python -m http.server 8000
```

Open `http://localhost:8000`. The page connects to `http://localhost:8080` when not hosted on GitHub Pages.

## API Testing

List recipes:

```bash
curl http://localhost:8080/api/v1/recipes
```

Create a recipe:

```bash
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Pancakes",
    "description": "Fluffy test pancakes",
    "categories": ["breakfast", "test"],
    "ingredients": ["1 cup flour", "2 eggs"],
    "steps": ["Mix ingredients", "Cook"],
    "links": [],
    "oven_temperature": null
  }'
```

## Admin Access

Writes (add/edit/delete recipes and images) require logging in with a PIN. Set in `backend/.env`:

```env
ADMIN_PIN=123456
AUTH_SECRET=
```

`ADMIN_PIN` must be at least 6 digits. In development both fall back to defaults (PIN `123456`), so you can leave them empty locally. In production (`APP_ENV=production`) both are required.

## Optional Image Storage

Image upload and delete endpoints require Cloudflare R2-compatible settings. Without them, the backend still runs, but image endpoints return `503`.

Set these variables when image storage is needed:

```env
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_PUBLIC_URL=
```

Leave all R2 variables empty to disable image storage. If any R2 variable is set, all R2 variables must be set.

## Common Commands

```bash
cd backend
go build ./...
go test ./...
go vet ./...
make run
make test
make test-coverage
```

Frontend validation is currently manual:

```bash
test -f frontend/index.html
test -f frontend/app.js
test -f frontend/api.js
node --check frontend/api.js
node --check frontend/app.js
```

Then load the page in a browser and test recipe CRUD, search/filtering, and image behavior if R2 is configured.

## Troubleshooting

| Issue | Check |
| --- | --- |
| CORS errors | Include the frontend origin in `CORS_ORIGIN` |
| Image upload returns `503` | R2 configuration is missing or incomplete |
| SQLite database locked | Run only one backend instance against the same SQLite file |
