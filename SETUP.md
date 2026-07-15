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

Edit `.env` before starting the server. `AUTH_USERNAME`, `AUTH_PASSWORD`, and `JWT_SECRET` are required; the app exits if any are empty.

```env
PORT=8080
APP_ENV=development
DATABASE_PATH=./recipes.db
CORS_ORIGIN=http://localhost:8080,http://localhost:8000
AUTH_USERNAME=admin
AUTH_PASSWORD=change-me
JWT_SECRET=replace-with-a-long-random-secret
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
```

## Frontend

```bash
cd frontend
python -m http.server 8000
```

Open `http://localhost:8000`. The page displays a login overlay and connects to `http://localhost:8080` when not hosted on GitHub Pages.

## Login And API Testing

Get a JWT:

```bash
TOKEN=$(
  curl -s http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"change-me"}' \
  | sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
)
```

List recipes:

```bash
curl http://localhost:8080/api/v1/recipes \
  -H "Authorization: Bearer $TOKEN"
```

Create a recipe:

```bash
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Authorization: Bearer $TOKEN" \
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
```

Then load the page in a browser and test login, recipe CRUD, search/filtering, and image behavior if R2 is configured.

## Troubleshooting

| Issue | Check |
| --- | --- |
| Backend exits on startup | Required auth variables are missing or empty |
| CORS errors | Include the frontend origin in `CORS_ORIGIN` |
| Unauthorized API responses | Log in first and send `Authorization: Bearer <token>` |
| Image upload returns `503` | R2 configuration is missing or incomplete |
| SQLite database locked | Run only one backend instance against the same SQLite file |
