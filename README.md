# Marellis Recipe Website

Marellis is a family recipe website for storing, browsing, and editing recipes. The project has a static vanilla JavaScript frontend and a Go REST API backed by SQLite.

## Current Features

- Sign-in with a single configured username and password
- JWT-protected recipe and category API
- Recipe create, read, update, and delete
- Category browsing and filtering
- Text search using SQL `LIKE` over recipe fields
- Optional oven temperature and external/recipe links
- Optional recipe image upload/delete through Cloudflare R2-compatible storage
- Static frontend suitable for GitHub Pages
- Backend deployment on Fly.io with SQLite on a mounted volume

## Architecture

- `frontend/` is a static site. It talks directly to the backend API from the browser.
- `backend/` is a Go API using Chi, sqlx, and SQLite.
- The backend runs ordered SQL migrations on startup and records applied versions in SQLite.
- The root `Dockerfile` is the Fly.io build path for the current root-level `fly.toml`.
- GitHub Actions validate the backend/frontend and deploy the frontend/backend from `master`.

## Technology Stack

- Frontend: HTML, CSS, vanilla JavaScript
- Backend: Go module version `1.24`, Chi, sqlx, modernc SQLite, go-playground/validator
- Auth: single shared username/password configured by environment, custom HS256 JWT middleware
- Storage: SQLite file; optional Cloudflare R2-compatible object storage for images
- Deployment: GitHub Pages for frontend, Fly.io for backend

## Repository Layout

```text
.
├── frontend/                 Static frontend
├── backend/                  Go API
│   ├── cmd/api/              API entry point
│   ├── internal/             Config, DB, handlers, middleware, models, repository, router
│   └── migrations/           SQL migration files
├── .github/workflows/        Validation and deployment workflows
├── Dockerfile                Current Fly.io backend Docker build from repo root
├── fly.toml                  Fly.io app configuration
├── SETUP.md                  Local setup details
├── DEPLOYMENT.md             Deployment details
├── AGENTS.md                 Codex repository instructions
├── CLAUDE.md                 Claude-specific instructions
└── TODO.md                   Proposed future work
```

## Quick Start

Backend:

```bash
cd backend
cp .env.example .env
# edit AUTH_USERNAME, AUTH_PASSWORD, JWT_SECRET, and CORS_ORIGIN as needed
go mod download
go run cmd/api/main.go
```

Frontend:

```bash
cd frontend
python -m http.server 8000
```

Open `http://localhost:8000`. The frontend uses `http://localhost:8080` outside GitHub Pages and `https://recipe-website-ai.fly.dev` on GitHub Pages.

## Configuration

Required backend variables:

- `AUTH_USERNAME`
- `AUTH_PASSWORD` or `AUTH_PASSWORD_HASH`
- `JWT_SECRET`

Common backend variables:

- `PORT` defaults to `8080`
- `APP_ENV` defaults to `development`
- `DATABASE_PATH` defaults to `./recipes.db`
- `CORS_ORIGIN` defaults to `http://localhost:8080` if unset

Optional R2 image storage variables:

- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET_NAME`
- `R2_PUBLIC_URL`

For production, prefer `AUTH_PASSWORD_HASH` with a bcrypt hash and leave `AUTH_PASSWORD` empty. Plain `AUTH_PASSWORD` remains supported for local development and backwards-compatible deployments.

## Validation

Backend:

```bash
cd backend
gofmt -w ./...
go build ./...
go test ./...
go vet ./...
```

Frontend:

```bash
test -f frontend/index.html
test -f frontend/app.js
test -f frontend/api.js
node --check frontend/api.js
node --check frontend/app.js
```

There is no frontend build step or package manager. Manual browser validation is currently expected for UI changes.

## API Summary

Public endpoints:

- `GET /health`
- `GET /ready`
- `POST /api/v1/auth/login`

JWT-protected endpoints:

- `GET /api/v1/recipes`
- `POST /api/v1/recipes`
- `GET /api/v1/recipes/{id}`
- `PUT /api/v1/recipes/{id}`
- `DELETE /api/v1/recipes/{id}`
- `POST /api/v1/recipes/{id}/image`
- `DELETE /api/v1/recipes/{id}/image`
- `GET /api/v1/categories`

See [backend/README.md](backend/README.md) for request examples.

## Deployment

- Backend: Fly.io uses the root `fly.toml` and root `Dockerfile`.
- Frontend: GitHub Pages deploys the contents of `frontend/`.
- GitHub Actions deploy backend and frontend only after the `Validate` workflow succeeds on `master`.

See [DEPLOYMENT.md](DEPLOYMENT.md) for details.

## Current Limitations

- Authentication is a single shared account configured through environment variables.
- JWTs are stored in browser `localStorage` and expire after 90 days.
- Search is SQL `LIKE` search, not SQLite FTS.
- Authentication is still a single shared account, not multi-user authorization.
- Database backup/restore is documented at a basic operational level, but should be rehearsed before relying on it for production recovery.

## Documentation

- [SETUP.md](SETUP.md) - local development setup
- [DEPLOYMENT.md](DEPLOYMENT.md) - deployment and CI/CD notes
- [backend/README.md](backend/README.md) - backend API and operations
- [frontend/README.md](frontend/README.md) - frontend structure and manual validation
- [TODO.md](TODO.md) - proposed future improvements
